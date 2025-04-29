package controller

import (
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

type AddressQueryParameters struct {
	Street      string
	HouseNumber string
	City        string
	ZipCode     string
}

func FetchOffers(c *gin.Context) {
	log.Warn("Not all parameter specified")
	c.String(200, "Hello World")
}
