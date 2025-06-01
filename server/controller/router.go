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
	r.Use(corsMiddleware())

	r.GET("/offers", FetchOffersByAddress)
	r.GET("/offers/shared/:shareId", FetchSharedOffers)
	r.POST("/offers/shared/:queryHash", ShareOffer)

	return r
}

// corsMiddleware sets up CORS headers for all routes
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Header("Access-Control-Expose-Headers", "Content-Length")
		c.Header("Access-Control-Allow-Credentials", "true")

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
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
	Provider       *string                `form:"provider"`
	Installation   *bool                  `form:"installation"`
	SpeedMin       *int                   `form:"speedMin"`
	Age            *int                   `form:"age"`
	CostMax        *int                   `form:"costMax"`
	ConnectionType *domain.ConnectionType `form:"connectionType"`
}

type OfferFilter func(domain.Offer) bool

func (filter FilterOptionParams) standardFilter(offer domain.Offer) bool {
	if filter.Provider != nil && *filter.Provider != "" && offer.Provider != *filter.Provider {
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
	if filter.ConnectionType != nil && *filter.ConnectionType != "" && offer.ConnectionType != *filter.ConnectionType {
		return false
	}
	return true
}

func (filter FilterOptionParams) isEmpty() bool {
	return (filter.Provider == nil || *filter.Provider == "") && filter.Installation == nil && filter.SpeedMin == nil &&
		filter.Age == nil && filter.CostMax == nil && (filter.ConnectionType == nil || *filter.ConnectionType == "")
}

func (filter FilterOptionParams) hash() string {
	agg := make([]byte, 0)
	if filter.Provider != nil {
		agg = fmt.Appendf(agg, "%s", *filter.Provider)
	}
	if filter.Installation != nil {
		agg = fmt.Appendf(agg, "%t", *filter.Installation)
	}
	if filter.SpeedMin != nil {
		agg = fmt.Appendf(agg, "%d", *filter.SpeedMin)
	}
	if filter.Age != nil {
		agg = fmt.Appendf(agg, "%d", *filter.Age)
	}
	if filter.CostMax != nil {
		agg = fmt.Appendf(agg, "%d", *filter.CostMax)
	}
	if filter.ConnectionType != nil {
		agg = fmt.Appendf(agg, "%s", (*filter.ConnectionType).String())
	}

	return string(utils.Hash(agg))
}

func FetchOffersByAddress(c *gin.Context) {
	now := time.Now().Unix()
	var userQuery domain.Query = domain.Query{
		Timestamp: now,
		Offers:    make(map[string]domain.Offer),
	}
	var addressQuery domain.Query = domain.Query{
		Timestamp: now,
		Offers:    make(map[string]domain.Offer),
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
	userQuery.GenerateAddressHash()
	addressQuery.Address = domain.Address{
		Street:      params.Street,
		HouseNumber: params.HouseNumber,
		City:        params.City,
		ZipCode:     params.ZipCode,
	}
	addressQuery.GenerateAddressHash()

	// save session id to user query
	userQuery.SessionID = params.SessionId

	// Set status for successful response
	c.Status(http.StatusOK)

	if queryJSON, err := json.Marshal(userQuery); err == nil {
		// Write the query information to the response
		fmt.Fprintf(c.Writer, "{\"query\": %s}\n", queryJSON)
		flusher.Flush()
	}

	ctx := c.Request.Context()
	combinedOfferChannel := make(chan domain.Offer)
	shouldApiRequest := true

	// retrieve cached offers for address
	cachedOffersInStream := make(chan struct{})
	if cachedQuery, _ := db.OfferCacheInstance.GetCachedQuery(ctx, addressQuery); cachedQuery != nil {
		log.Debugf("Found cached query for address %s", addressQuery.Address)
		shouldApiRequest = now-cachedQuery.Timestamp > utils.Cfg.Server.FreshnessWindowSec

		go func() {
			for _, offer := range cachedQuery.Offers {
				// if a new request gonna happen, set preliminary flag to true to indicate that these are cached and not live from api
				offer.HelperIsPreliminary = shouldApiRequest
				combinedOfferChannel <- offer
			}
			close(cachedOffersInStream)
		}()
	} else {
		close(cachedOffersInStream)
	}

	var offersStreamingDone <-chan struct{}
	if shouldApiRequest {
		// Start the streaming service
		liveOffersPubSubChannel, errChannel := offerService.FetchOffersStream(ctx, addressQuery.Address)
		// Process errors
		go func() {
			for {
				select {
				case err, ok := <-errChannel:
					if !ok {
						return
					}
					log.WithError(err).Warn("Error while fetching offers")
				case <-ctx.Done():
					// Context cancelled, stop processing
					return
				}
			}
		}()
		// save all live offers in address cache so that if multiple users with different filters request the same address, they can use cached offers
		dumpChan, addressCacheDone := cacheOffers(ctx, &addressQuery, liveOffersPubSubChannel.Subscribe(), db.OfferCacheInstance.CacheQuery)
		utils.DumpChannel(dumpChan)

		// put live offers into combined stream to stream to output
		liveOffersInStream := make(chan struct{})
		go func(liveOffersChannel <-chan domain.Offer) {
			for {
				select {
				case offer, ok := <-liveOffersChannel:
					if !ok {
						// all offers are processed, close the channel to signal all live offers are in streaming channel
						close(liveOffersInStream)
						return
					}

					combinedOfferChannel <- offer
				case <-ctx.Done():
					// Context cancelled, stop processing
					close(liveOffersInStream)
					return
				}
			}
		}(liveOffersPubSubChannel.Subscribe())

		// cache offers for user which are preliminary and live to ensure share links with both contained
		userCachedOfferChannel, _ := cacheOffers(ctx, &userQuery, combinedOfferChannel, db.UserOfferCacheInstance.CacheQuery)

		// stream everything that is cached for later sharing to the user
		offersStreamingDone = handleOfferStreaming(ctx, c.Writer, flusher, userCachedOfferChannel)

		// wait until all cached offers are in streaming channel
		<-cachedOffersInStream
		log.Debug("Cached offers in combined stream")
		// wait until all live offers are in streaming channel
		<-liveOffersInStream
		log.Debug("Live offers in combined stream")

		close(combinedOfferChannel)

		// wait until all offers cached for address before closing request
		<-addressCacheDone
		log.Debug("caching offers for address done")
	} else {
		log.Debug("Using cached offers for address, no new API request will be made")

		// offers by cache are counted as valid as no new api request is made
		// therefore they need to be saved in the user cache
		cachedOffers, _ := cacheOffers(ctx, &userQuery, combinedOfferChannel, db.UserOfferCacheInstance.CacheQuery)
		offersStreamingDone = handleOfferStreaming(ctx, c.Writer, flusher, cachedOffers)

		// wait until cached offers are all in streaming channel
		<-cachedOffersInStream
		log.Debug("Cached offers in combined stream")
		close(combinedOfferChannel)
	}

	// Wait for the streaming
	<-offersStreamingDone

	// write query related information to the response
	log.Debug("Stream successfully closed\n")
}

func ShareOffer(c *gin.Context) {
	queryHash := c.Param("queryHash")
	if queryHash == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query hash is required"})
		return
	}

	sessionId := c.Query("sessionId")
	if sessionId == "" {
		log.Warn("Session ID not specified")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing session ID"})
		return
	}

	// as we filter client side, but want to display the same offers in the share link, we need to filter the cached offers now before creating the snapshot
	var filterParams FilterOptionParams
	if err := c.ShouldBind(&filterParams); err != nil {
		log.WithError(err).Warn("Failed to parse filter query parameters")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid filter query parameters"})
		return
	}
	log.Debugf("Filter parameters: %+v", filterParams)

	query, err := db.UserOfferCacheInstance.GetCachedUserQuery(c.Request.Context(), queryHash+":"+sessionId)
	if err != nil {
		log.WithError(err).Error("Failed to retrieve cached query for sharing")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve cached query"})
		return
	}

	// filter offers based on the provided filter parameters
	isFilterEmpty := filterParams.isEmpty()
	filteredOffers := make(map[string]domain.Offer)

	// create shareId by hashing of offer hashes, filterParams and queryHash
	idAgg := make([]byte, 0)
	idAgg = fmt.Appendf(idAgg, "%s%s", queryHash, filterParams.hash())
	for _, offer := range query.Offers {
		if isFilterEmpty || filterParams.standardFilter(offer) {
			filteredOffers[offer.HelperOfferHash] = offer
			idAgg = fmt.Appendf(idAgg, "%s%t", offer.HelperOfferHash, offer.HelperIsPreliminary)
		}
	}
	query.Offers = filteredOffers

	if len(query.Offers) == 0 {
		log.Warn("Can not share empty query")
		c.JSON(http.StatusNotFound, gin.H{"error": "Cannot share an empty query"})
		return
	}

	shareId := utils.HashURLEncoded([]byte(idAgg))

	exists, err := db.QueryExists(c.Request.Context(), shareId)
	if err != nil {
		return
	}

	if exists {
		c.JSON(http.StatusOK, gin.H{"shareId": shareId})
		log.Infof("Query already shared: %s", shareId)
		return
	}

	queryEntity := db.QueryEntity{
		ShareId: shareId,
		Query:   *query,
	}
	// save query in database for sharing
	shareId, err = db.SaveQuery(c.Request.Context(), queryEntity)
	if err != nil {
		log.WithError(err).Error("Failed to save query for sharing")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save query for sharing"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"shareId": shareId})
	log.Infof("Query shared successfully with ID: %s", shareId)
}

func FetchSharedOffers(c *gin.Context) {
	shareId := c.Param("shareId")
	if shareId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Share ID is required"})
		return
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

	query, err := db.GetQueryById(c.Request.Context(), shareId)
	if err != nil {
		log.WithError(err).Error("Failed to retrieve shared query")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve shared query"})
		return
	}

	if query == nil {
		log.Warnf("No query found for share ID %s", shareId)
		c.JSON(http.StatusNotFound, gin.H{"error": "No query found for the provided share ID"})
		return
	}

	// Set status for successful response
	c.Status(http.StatusOK)

	offers := query.Offers
	query.Offers = nil
	if queryJSON, err := json.Marshal(query); err == nil {
		// Write the query information to the response
		fmt.Fprintf(c.Writer, "{\"query\": %s}\n", queryJSON)
		flusher.Flush()
	}

	for _, offer := range offers {
		if offerJSON, err := json.Marshal(offer); err == nil {
			fmt.Fprintf(c.Writer, "{\"offer\": %s}\n", offerJSON)
			flusher.Flush()
		}
	}
}

func cacheOffers(ctx context.Context, query *domain.Query, offersChannel <-chan domain.Offer, cacheFunc func(ctx context.Context, query domain.Query) error) (<-chan domain.Offer, <-chan struct{}) {
	done := make(chan struct{})
	cachedOffersChannel := make(chan domain.Offer)

	go func() {
		for {
			select {
			case offer, ok := <-offersChannel:
				if !ok {
					// Cache the offers for the address
					log.Debugf("Caching %d offers", len(query.Offers))
					if err := cacheFunc(ctx, *query); err != nil {
						log.WithError(err).Error("Failed to cache offers for address")
					}
					close(cachedOffersChannel)
					close(done)
					return
				}
				if offer.HelperOfferHash == "" {
					// Generate hash for the offer if not already set
					offer.GenerateHash()
				}

				offerInQuery, exists := query.Offers[offer.HelperOfferHash]
				if !exists || offerInQuery.HelperIsPreliminary {
					// Send the offer to the fanout channel
					cachedOffersChannel <- offer

					// Also append the offer to the address query for caching
					query.Offers[offer.HelperOfferHash] = offer
				}
			case <-ctx.Done():
				// Context cancelled, stop processing
				log.Debug("Context cancelled, stopping offer caching")
				close(cachedOffersChannel)
				close(done)
				return
			}
		}
	}()

	return cachedOffersChannel, done
}

func handleOfferStreaming(c context.Context, writer io.Writer, flusher http.Flusher, offersChannel <-chan domain.Offer) (done chan struct{}) {
	done = make(chan struct{})

	go func() {
		for {
			select {
			case offer, ok := <-offersChannel:
				if !ok {
					close(done)
					return
				}

				if offerJSON, err := json.Marshal(offer); err == nil {
					fmt.Fprintf(writer, "{\"offer\": %s}\n", offerJSON)
					flusher.Flush()
				} else {
					log.WithError(err).Warn("Failed to marshal offer")
				}

			case <-c.Done():
				// Context cancelled, stop processing
				log.Debug("Context cancelled, stopping offer streaming")
				close(done)
				return
			}
		}
	}()

	return done
}
