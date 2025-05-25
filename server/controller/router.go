package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"server/domain"
	"server/service"

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

type AddressQueryParameters struct {
	Street      string `form:"street"`
	HouseNumber string `form:"houseNumber"`
	City        string `form:"city"`
	ZipCode     string `form:"plz"`
}

func FetchOffersByAddress(c *gin.Context) {
	// Parse address parameters from query
	var params AddressQueryParameters
	if err := c.ShouldBindQuery(&params); err != nil {
		log.WithError(err).Warn("Failed to parse query parameters")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters"})
		return
	}

	// Validate required parameters
	if params.Street == "" || params.HouseNumber == "" || params.City == "" || params.ZipCode == "" {
		log.Warn("Not all parameters specified")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required address parameters"})
		return
	}

	// Create address object
	address := domain.Address{
		Street:      params.Street,
		HouseNumber: params.HouseNumber,
		City:        params.City,
		ZipCode:     params.ZipCode,
	}

	// Set response headers for streaming
	c.Header("Content-Type", "application/json")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// Set status for successful response
	c.Status(http.StatusOK)

	writer := c.Writer
	flusher, ok := writer.(http.Flusher)
	if !ok {
		log.Warn("Writer doesn't support flushing")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Streaming not supported"})
		return
	}

	// Create channels for streaming offers
	offersChannel := make(chan domain.Offer)
	errChannel := make(chan error)

	ctx := c.Request.Context()
	// retrieve cached offers
	cachedOffers := offerService.GetOffersCached(ctx, address)
	if cachedOffers != nil {
		log.WithField("count", len(cachedOffers)).Debug("Retrieved cached offers")
	}
	for _, offer := range cachedOffers {
		offer.HelperOfferStatus = domain.OfferStatus(domain.Preliminary)
		if offerJSON, err := json.Marshal(offer); err == nil {
			fmt.Fprintf(writer, "offer: %s\n\n", offerJSON)
			flusher.Flush()
		}
	}

	// Start the streaming service and get the completion signal
	offersFetchingDone := offerService.FetchOffersStream(ctx, address, offersChannel, errChannel)
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

	offersStreamingDone, offersForCacheChannel := handleOfferStreaming(c, writer, flusher, offersChannel)

	offersCachingDone := make(chan struct{})
	// Process errors
	go func() {
		validOffers := make([]domain.Offer, 0)

		for {
			select {
			case offer, ok := <-offersForCacheChannel:
				if !ok {
					log.Debugf("Caching %d valid offers", len(validOffers))
					offerService.CacheOffers(ctx, address, validOffers)
					close(offersCachingDone)
					return
				}

				validOffers = append(validOffers, offer)
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

	// write query related information to the response
	log.Debug("Stream successfully closed")
}

func handleOfferStreaming(c *gin.Context, writer gin.ResponseWriter, flusher http.Flusher, offersChannel <-chan domain.Offer) (done chan struct{}, fanOffersChannel chan domain.Offer) {
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
				offer.HelperOfferStatus = domain.OfferStatus(domain.Valid)
				if offerJSON, err := json.Marshal(offer); err == nil {
					fmt.Fprintf(writer, "offer: %s\n\n", offerJSON)
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

func FetchSharedOffers(c *gin.Context) {

	c.JSON(http.StatusNotImplemented, gin.H{"error": "Shared offers not implemented yet"})
}
