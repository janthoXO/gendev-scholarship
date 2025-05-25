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
	r.GET("/offers/shared", FetchSharedOffers)
	// TODO add a post endpoint which buys an offer. The request gets put into a queue until the api is available again
	// r.POST("/offers")

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

func FetchOffersByAddress(c *gin.Context) {
	var query domain.Query = domain.Query{
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
	query.Address = domain.Address{
		Street:      params.Street,
		HouseNumber: params.HouseNumber,
		City:        params.City,
		ZipCode:     params.ZipCode,
	}
	query.SessionID = params.SessionId

	// Set status for successful response
	c.Status(http.StatusOK)

	// Create channels for streaming offers
	offersChannel := make(chan domain.Offer)
	errChannel := make(chan error)

	ctx := c.Request.Context()
	// retrieve cached offers
	cachedQuery, _ := db.OfferCacheInstance.GetCachedQuery(ctx, query)
	if cachedQuery != nil {
		log.Debugf("Found cached query for address %s", query.Address)
		for _, offer := range cachedQuery.Offers {
			offer.HelperIsPreliminary = true
			if offerJSON, err := json.Marshal(offer); err == nil {
				fmt.Fprintf(c.Writer, "{offer: %s}\n", offerJSON)
				flusher.Flush()
			}
		}
	}

	// Start the streaming service and get the completion signal
	offersFetchingDone := offerService.FetchOffersStream(ctx, query.Address, offersChannel, errChannel)
	go func() {
		select {
		case <-offersFetchingDone:
			log.Debug("Offers fetching completed")
			close(offersChannel)
			close(errChannel)
		case <-c.Request.Context().Done():
			log.Debug("Client disconnected while fetching offers")
		}
	}()

	offersStreamingDone, offersForCacheChannel := handleOfferStreaming(c, flusher, offersChannel)

	offersCachingDone := make(chan struct{})
	// Process errors
	go func() {
		for {
			select {
			case offer, ok := <-offersForCacheChannel:
				if !ok {
					db.OfferCacheInstance.CacheQuery(ctx, query)
					db.UserOfferCacheInstance.CacheQuery(ctx, query)
					close(offersCachingDone)
					return
				}

				query.Offers = append(query.Offers, offer)
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

	// Wait for the streaming and caching to finish
	<-offersStreamingDone
	<-offersCachingDone

	// set offers to nil to not send them again
	query.Offers = nil
	if queryJSON, err := json.Marshal(query); err == nil {
		fmt.Fprintf(c.Writer, "{query: %s}\n", queryJSON)
		flusher.Flush()
	}

	// write query related information to the response
	log.Debug("Stream successfully closed")
}

func FetchSharedOffers(c *gin.Context) {

	c.JSON(http.StatusNotImplemented, gin.H{"error": "Shared offers not implemented yet"})
}

func handleOfferStreaming(c *gin.Context, flusher http.Flusher, offersChannel <-chan domain.Offer) (done chan struct{}, fanOffersChannel chan domain.Offer) {
	done = make(chan struct{})
	fanOffersChannel = make(chan domain.Offer)

	go func() {
		for {
			select {
			case offer, ok := <-offersChannel:
				if !ok {
					close(fanOffersChannel)
					close(done)
					return
				}

				fanOffersChannel <- offer
				offer.HelperIsPreliminary = false
				if offerJSON, err := json.Marshal(offer); err == nil {
					fmt.Fprintf(c.Writer, "{offer: %s}\n", offerJSON)
					flusher.Flush()
				} else {
					log.WithError(err).Warn("Failed to marshal offer")
				}

			case <-c.Request.Context().Done():
				return
			}
		}
	}()

	return done, fanOffersChannel
}
