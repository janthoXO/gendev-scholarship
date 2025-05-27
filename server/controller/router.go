package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"server/db"
	"server/domain"
	"server/service"
	"server/utils"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func SetupRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.ErrorLogger())
	r.Use(gin.Recovery())

	r.GET("/offers", FetchOffersByAddress)
	r.GET("/offers/:queryHash/shared", FetchSharedOffers)
	r.POST("/offers/:queryHash/shared", ShareOffer)

	return r
}

var offerService = service.OfferServiceImpl{}

type FetchOffersQueryParameters struct {
	Street      string `form:"street"`
	HouseNumber string `form:"houseNumber"`
	City        string `form:"city"`
	ZipCode     string `form:"plz"`
	SessionId   string `form:"sessionId"`
}

type FilterOptionParams struct {
	Provider       *string `form:"provider"`
	Installation   *bool   `form:"installation"`
	SpeedMin       *int    `form:"speedMin"`
	Age            *int    `form:"age"`
	CostMax        *int    `form:"costMax"`
	ConnectionType *string `form:"connectionType"`
}

type OfferFilter func(domain.Offer) bool

func (filter FilterOptionParams) standardFilter(offer domain.Offer) bool {
	if filter.Provider != nil && offer.Provider != *filter.Provider {
		return false
	}
	if filter.Installation != nil && offer.InstallationService != *filter.Installation {
		return false
	}
	if filter.SpeedMin != nil && offer.Speed < *filter.SpeedMin {
		return false
	}
	if filter.Age != nil && offer.MaxAgePerson < *filter.Age {
		return false
	}
	if filter.CostMax != nil && offer.MonthlyCostInCent > *filter.CostMax {
		return false
	}
	if filter.ConnectionType != nil && offer.ConnectionType != *filter.ConnectionType {
		return false
	}
	return true
}

func (filter FilterOptionParams) isEmpty() bool {
	return filter.Provider == nil && filter.Installation == nil && filter.SpeedMin == nil &&
		filter.Age == nil && filter.CostMax == nil && filter.ConnectionType == nil
}

func FetchOffersByAddress(c *gin.Context) {
	now := time.Now().Unix()
	var userQuery domain.Query = domain.Query{
		Timestamp: now,
	}
	var addressQuery domain.Query = domain.Query{
		Timestamp: now,
	}

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		log.Warn("Writer doesn't support flushing")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Streaming not supported"})
		return
	}

	// Set response headers for streaming
	// Set headers for NDJSON
	c.Header("Content-Type", "application/x-ndjson")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "close") // Will close when done
	c.Header("Access-Control-Allow-Origin", "*")

	// Parse address parameters from query
	var params FetchOffersQueryParameters
	if err := c.ShouldBindQuery(&params); err != nil {
		log.WithError(err).Warn("Failed to parse query parameters")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters"})
		return
	}

	// Validate required parameters
	if params.Street == "" || params.HouseNumber == "" || params.City == "" || params.ZipCode == "" {
		log.Warn("Not all address parameters specified")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required address parameters"})
		return
	}
	if params.SessionId == "" {
		log.Warn("Session ID not specified")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing session ID"})
		return
	}

	// Create address object
	userQuery.Address = domain.Address{
		Street:      params.Street,
		HouseNumber: params.HouseNumber,
		City:        params.City,
		ZipCode:     params.ZipCode,
	}
	addressQuery.Address = domain.Address{
		Street:      params.Street,
		HouseNumber: params.HouseNumber,
		City:        params.City,
		ZipCode:     params.ZipCode,
	}

	// save session id to user query
	userQuery.SessionID = params.SessionId

	var filterParams FilterOptionParams
	if err := c.ShouldBindQuery(&filterParams); err != nil {
		log.WithError(err).Warn("Failed to parse filter query parameters")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid filter query parameters"})
		return
	}

	// Set status for successful response
	c.Status(http.StatusOK)

	ctx := c.Request.Context()
	streamingChannel := make(chan domain.Offer)
	shouldApiRequest := true

	// retrieve cached offers for address
	cachedOffersInStream := make(chan struct{})
	cachedQuery, _ := db.OfferCacheInstance.GetCachedQuery(ctx, addressQuery)
	if cachedQuery != nil {
		log.Debugf("Found cached query for address %s", addressQuery.Address)
		shouldApiRequest = now-cachedQuery.Timestamp > utils.Cfg.Server.ApiCooldownSec

		go func() {
			for _, offer := range cachedQuery.Offers {
				// Apply filter to all offers
				if !filterParams.standardFilter(offer) {
					continue
				}

				// if a new request gonna happen, set preliminary flag to true to indicate that these are cached and not live from api
				offer.HelperIsPreliminary = shouldApiRequest
				streamingChannel <- offer
			}
			log.Debug("Cached offers streaming done")
			close(cachedOffersInStream)
		}()
	} else {
		close(cachedOffersInStream)
	}

	var offersStreamingDone <-chan struct{}
	if shouldApiRequest {
		// Start the streaming service
		offersChannel, errChannel := offerService.FetchOffersStream(ctx, addressQuery.Address)
		// Process errors
		go func() {
			for {
				select {
				case err, ok := <-errChannel:
					if !ok {
						continue
					}
					log.WithError(err).Warn("Error while fetching offers")

				case <-ctx.Done():
					// Context cancelled, stop processing
					return
				}
			}
		}()
		userOfferChannel, _ := cacheOffers(ctx, addressQuery, offersChannel, db.OfferCacheInstance.CacheQuery, nil)
		filteredOffersChannel, _ := cacheOffers(ctx, userQuery, userOfferChannel, db.UserOfferCacheInstance.CacheQuery, filterParams.standardFilter)

		liveOffersInStream := make(chan struct{})
		go func() {
			for {
				select {
				case offer, ok := <-filteredOffersChannel:
					if !ok {
						// all offers are processed, close the channel to signal all live offers are in streaming channel
						close(liveOffersInStream)
						log.Debug("Live offers streaming done")
						return
					}

					// set preliminary flag to false to indicate that these are live from api
					offer.HelperIsPreliminary = false
					streamingChannel <- offer
				case <-ctx.Done():
					// Context cancelled, stop processing
					close(streamingChannel)
					return
				}
			}
		}()

		offersStreamingDone = handleOfferStreaming(ctx, c.Writer, flusher, streamingChannel, nil)

		// wait until all cached offers are in streaming channel
		<-cachedOffersInStream
		// wait until all live offers are in streaming channel
		<-liveOffersInStream
		close(streamingChannel)
	} else {
		log.Debug("Using cached offers for address, no new API request will be made")

		// offers by cache are counted as valid as no new api request is made
		// therefore they need to be saved in the user cache
		cachedOffers, _ := cacheOffers(ctx, userQuery, streamingChannel, db.UserOfferCacheInstance.CacheQuery, nil)
		offersStreamingDone = handleOfferStreaming(ctx, c.Writer, flusher, cachedOffers, nil)

		// wait until cached offers are all in streaming channel
		<-cachedOffersInStream
		close(streamingChannel)
	}

	// Wait for the streaming
	<-offersStreamingDone

	// set offers to nil to not send them again
	userQuery.Offers = nil
	if queryJSON, err := json.Marshal(userQuery); err == nil {
		// Write the query information to the response
		fmt.Fprintf(c.Writer, "{\"query\": %s}\n", queryJSON)
		flusher.Flush()
	}

	// write query related information to the response
	log.Debug("Stream successfully closed\n")
}

func ShareOffer(c *gin.Context) {
	queryHash := c.Param("queryHash")
	if queryHash == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query hash is required"})
		return
	}
	query, err := db.UserOfferCacheInstance.GetCachedUserQuery(c.Request.Context(), queryHash)
	if err != nil {
		log.WithError(err).Error("Failed to retrieve cached query for sharing")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve cached query"})
		return
	}

	// TODO save query in database for sharing

	c.JSON(http.StatusOK, gin.H{"link": fmt.Sprintf("offers/%s/shared?street=%s&houseNumber=%s&city=%s&plz=%s", queryHash, query.Address.Street, query.Address.HouseNumber, query.Address.City, query.Address.ZipCode)})
}

func FetchSharedOffers(c *gin.Context) {

	c.JSON(http.StatusNotImplemented, gin.H{"error": "Shared offers not implemented yet"})
}

func cacheOffers(ctx context.Context, query domain.Query, offersChannel <-chan domain.Offer, cacheFunc func(ctx context.Context, query domain.Query) error, filter OfferFilter) (<-chan domain.Offer, <-chan struct{}) {
	done := make(chan struct{})
	cachedOffersChannel := make(chan domain.Offer)

	go func() {
		for {
			select {
			case offer, ok := <-offersChannel:
				if !ok {
					// Cache the offers for the address
					if err := cacheFunc(ctx, query); err != nil {
						log.WithError(err).Error("Failed to cache offers for address")
					}
					close(cachedOffersChannel)
					close(done)
					return
				}
				if filter == nil || filter(offer) {
					// Send the offer to the fanout channel
					cachedOffersChannel <- offer
					// Also append the offer to the address query for caching
					query.Offers = append(query.Offers, offer)
				}
			case <-ctx.Done():
				// Context cancelled, stop processing
				close(cachedOffersChannel)
				close(done)
				return
			}
		}
	}()

	return cachedOffersChannel, done
}

func handleOfferStreaming(c context.Context, writer io.Writer, flusher http.Flusher, offersChannel <-chan domain.Offer, filter OfferFilter) (done chan struct{}) {
	done = make(chan struct{})

	go func() {
		for {
			select {
			case offer, ok := <-offersChannel:
				if !ok {
					close(done)
					return
				}

				// If a filter is provided, only send the offer to the client if it matches the filter
				if filter == nil || filter(offer) {
					if offerJSON, err := json.Marshal(offer); err == nil {
						fmt.Fprintf(writer, "{\"offer\": %s}\n", offerJSON)
						flusher.Flush()
					} else {
						log.WithError(err).Warn("Failed to marshal offer")
					}
				}

			case <-c.Done():
				// Context cancelled, stop processing
				close(done)
				return
			}
		}
	}()

	return done
}
