package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"server/db"
	"server/domain"
	"server/service"
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
	var userQuery domain.Query = domain.Query{
		Timestamp: time.Now(),
	}
	var addressQuery domain.Query = domain.Query{
		Timestamp: time.Now(),
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

	// retrieve cached offers for address
	cachedOffersStreamingDone := make(chan struct{})
	go func() {
		cachedQuery, _ := db.OfferCacheInstance.GetCachedQuery(ctx, addressQuery)
		if cachedQuery != nil {
			log.Debugf("Found cached query for address %s", addressQuery.Address)
			for _, offer := range cachedQuery.Offers {
				// Apply filter to all offers
				if !filterParams.standardFilter(offer) {
					continue
				}

				// set preliminary flag to true to indicate that these are cached and not live from api
				offer.HelperIsPreliminary = true
				if offerJSON, err := json.Marshal(offer); err == nil {
					fmt.Fprintf(c.Writer, "{offer: %s}\n", offerJSON)
					flusher.Flush()
				}
			}
		}
		log.Debug("Cached offers streaming done")
		close(cachedOffersStreamingDone)
	}()

	// Start the streaming service
	offersChannel, errChannel := offerService.FetchOffersStream(ctx, addressQuery.Address)

	filteredOffersChannel := make(chan domain.Offer)
	// Process errors
	go func() {
		for {
			select {
			case offer, ok := <-offersChannel:
				if !ok {
					// cache query with all offers for address
					// cache query for user with filtered offers
					db.OfferCacheInstance.CacheQuery(ctx, addressQuery)
					db.UserOfferCacheInstance.CacheQuery(ctx, userQuery)
					close(filteredOffersChannel)
					return
				}

				if filterParams.standardFilter(offer) {
					userQuery.Offers = append(userQuery.Offers, offer)
					filteredOffersChannel <- offer
				}
				addressQuery.Offers = append(addressQuery.Offers, offer)
			case err, ok := <-errChannel:
				if !ok {
					continue
				}
				log.WithError(err).Warn("Error while fetching offers")

			case <-c.Request.Context().Done():
				// Signal that we've finished writing the stream
				return
			}
		}
	}()

	// handle streaming of offers
	offersStreamingDone := handleOfferStreaming(c, flusher, filteredOffersChannel, func(o domain.Offer) bool { return true })

	// Wait for the streaming
	<-cachedOffersStreamingDone
	<-offersStreamingDone

	// set offers to nil to not send them again
	userQuery.Offers = nil
	if queryJSON, err := json.Marshal(userQuery); err == nil {
		// Write the query information to the response
		fmt.Fprintf(c.Writer, "{query: %s}\n", queryJSON)
		flusher.Flush()
	}

	// write query related information to the response
	log.Debug("Stream successfully closed")
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

func handleOfferStreaming(c *gin.Context, flusher http.Flusher, offersChannel <-chan domain.Offer, filter OfferFilter) (done chan struct{}) {
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
					offer.HelperIsPreliminary = false
					if offerJSON, err := json.Marshal(offer); err == nil {
						fmt.Fprintf(c.Writer, "{offer: %s}\n", offerJSON)
						flusher.Flush()
					} else {
						log.WithError(err).Warn("Failed to marshal offer")
					}
				}

			case <-c.Request.Context().Done():
				return
			}
		}
	}()

	return done
}
