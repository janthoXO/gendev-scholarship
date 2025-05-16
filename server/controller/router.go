package controller

import (
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

	allOffers, err := offerService.FetchOffers(address)
	if err != nil {
		log.WithError(err).Warn("Failed to fetch offers")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch offers"})
		return
	}

	// Handle the case when no offers are available
	if len(allOffers) == 0 {
		log.Warn("No offers available from any provider")
	}

	// Return all offers
	c.JSON(http.StatusOK, allOffers)
}
