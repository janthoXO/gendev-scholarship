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

	// Start the streaming service (it will close the channels when done)
	offerService.FetchOffersStream(c.Request.Context(), address, offersChannel, errChannel)

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
	processingComplete := make(chan bool)

	// Process offers and errors
	go func() {
		for {
			select {
			case offer, ok := <-offersChannel:
				if !ok {
					// Channel was closed, which means all providers are done
					processingComplete <- true
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
					continue
				}
				log.WithError(err).Warn("Error while fetching offers")
			}
		}
	}()

	// Wait for processing completion or client disconnect
	select {
	case <-processingComplete:
		// Processing completed (offers channel was closed)
		log.Info("Processing completed - all providers finished")
		writer.Write([]byte("]"))
		flusher.Flush()
	case <-c.Request.Context().Done():
		log.Info("Client disconnected")
		return
	}
}
