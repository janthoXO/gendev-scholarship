package controller

import (
	"encoding/json"
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

	r.GET("/offers", FetchOffers)
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

func FetchOffers(c *gin.Context) {
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

	// Create channels for streaming offers
	offersChannel := make(chan domain.Offer)
	errChannel := make(chan error)

	// Start the streaming service and get the completion signal
	completionSignal := offerService.FetchOffersStream(c.Request.Context(), address, offersChannel, errChannel)

	// Stream offers to the client
	writer := c.Writer
	flusher, ok := writer.(http.Flusher)
	if !ok {
		log.Warn("Writer doesn't support flushing")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Streaming not supported"})
		return
	}

	// Stream the response as JSON array manually
	writer.Write([]byte("["))
	flusher.Flush()

	first := true

	// Create a separate done channel for the streaming goroutine
	streamingDone := make(chan struct{})

	// Process offers and errors
	go func() {
		for {
			select {
			case offer, ok := <-offersChannel:
				if !ok {
					// offersChannel closed, exit the loop
					log.Debug("Offers channel closed, finishing stream")
					writer.Write([]byte("]"))
					flusher.Flush()
					// Signal that we've finished writing the stream
					close(streamingDone)
					return
				}

				// Add comma separator if not the first offer
				if !first {
					writer.Write([]byte(","))
				} else {
					first = false
				}

				// Marshal and write the offer
				if offerJSON, err := json.Marshal(offer); err == nil {
					writer.Write(offerJSON)
					flusher.Flush()
				} else {
					log.WithError(err).Warn("Failed to marshal offer")
				}

			case err, ok := <-errChannel:
				if !ok {
					// errChannel closed but continue waiting for offers
					continue
				}
				log.WithError(err).Warn("Error while fetching offers")

			case <-c.Request.Context().Done():
				// Signal that we've finished writing the stream
				close(streamingDone)
				return
			}
		}
	}()

	// Wait for providers completion, processing completion or client disconnect
	select {
	case <-completionSignal:
		// All providers have completed
		log.Debug("All providers have completed")

		// Close the channels as the router is responsible for them
		close(offersChannel)
		close(errChannel)

		// Wait for the streaming goroutine to finish writing
		<-streamingDone
		log.Debug("Stream successfully closed")

	case <-c.Request.Context().Done():
		log.Debug("Client disconnected")
		return
	}
}
