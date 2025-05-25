package db

import (
	"context"
	"encoding/json"
	"fmt"
	"server/domain"
	"server/utils"
	"time"

	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
)

type offerCache struct {
	redisClient *redis.Client
}

var (
	OfferCacheInstance offerCache
)

// Initialize the Redis client
func InitOfferCache() {
	OfferCacheInstance = offerCache{}
	OfferCacheInstance.redisClient = redis.NewClient(&redis.Options{
		Addr:     utils.Cfg.OfferCache.Url,      // Redis server address
		Password: utils.Cfg.OfferCache.Password, // No password set
		DB:       0,                             // Use default DB
	})

	// Test connection
	ctx := context.Background()
	if err := OfferCacheInstance.redisClient.Ping(ctx).Err(); err != nil {
		log.WithError(err).Warn("Failed to connect to Offer Redis")
	} else {
		log.Info("Connected to Offer Redis successfully")
	}
}

// CacheKey generates a cache key from an address
func (cache offerCache) cacheKey(query domain.Query) string {
	if query.HelperAddressHash == "" {
		query.GenerateAddressHash()
	}

	return query.HelperAddressHash
}

// GetCachedQuery retrieves a cached query for an address from the cache
func (cache offerCache) GetCachedQuery(ctx context.Context, query domain.Query) (*domain.Query, error) {
	key := cache.cacheKey(query)
	data, err := cache.redisClient.Get(ctx, key).Bytes()
	if err != nil {
		if err != redis.Nil {
			log.WithError(err).Warn("Failed to get data from Redis")
			return nil, fmt.Errorf("failed to get data from Redis: %w", err)
		}

		return nil, nil
	}

	if err := json.Unmarshal(data, &query); err != nil {
		log.WithError(err).Error("Failed to unmarshal cached query")
		return nil, fmt.Errorf("failed to unmarshal cached query: %w", err)
	}

	log.Debug("Retrieved query from offer cache")
	return &query, nil
}

// CacheQuery stores a query for an address in the cache
func (cache offerCache) CacheQuery(ctx context.Context, query domain.Query) error {
	data, err := json.Marshal(query)
	if err != nil {
		log.WithError(err).Error("Failed to marshal query for caching")
		return fmt.Errorf("failed to marshal query: %w", err)
	}

	key := cache.cacheKey(query)
	// Cache for 1 day
	if err := cache.redisClient.Set(ctx, key, data, time.Hour*24).Err(); err != nil {
		log.WithError(err).Error("Failed to store query in Redis")
	}

	log.Debugf("Stored query in offer cache with key: %s", key)
	return nil
}
