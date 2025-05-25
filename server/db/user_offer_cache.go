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

type userOfferCache struct {
	redisClient *redis.Client
}

var (
	UserOfferCacheInstance userOfferCache
)

// Initialize the Redis client
func InitUserOfferCache() {
	UserOfferCacheInstance = userOfferCache{}
	UserOfferCacheInstance.redisClient = redis.NewClient(&redis.Options{
		Addr:     utils.Cfg.UserOfferCache.Url,      // Redis server address
		Password: utils.Cfg.UserOfferCache.Password, // No password set
		DB:       0,                                 // Use default DB
	})

	// Test connection
	ctx := context.Background()
	if err := UserOfferCacheInstance.redisClient.Ping(ctx).Err(); err != nil {
		log.WithError(err).Warn("Failed to connect to User-Offer Redis")
	} else {
		log.Info("Connected to User-Offer Redis successfully")
	}
}

// CacheKey generates a cache key from an address
func (cache userOfferCache) cacheKey(query domain.Query) string {
	if query.HelperAddressHash == "" {
		query.GenerateAddressHash()
	}

	return query.HelperAddressHash + ":" + query.SessionID
}

// GetCachedUserQuery retrieves a cached query for a user from the cache
func (cache userOfferCache) GetCachedUserQueryByQuery(ctx context.Context, query domain.Query) (*domain.Query, error) {
	key := cache.cacheKey(query)
	return cache.GetCachedUserQuery(ctx, key)
}

// GetCachedUserQuery retrieves a cached query for a user from the cache
func (cache userOfferCache) GetCachedUserQuery(ctx context.Context, key string) (*domain.Query, error) {
	var query domain.Query
	data, err := UserOfferCacheInstance.redisClient.Get(ctx, key).Bytes()
	if err != nil {
		if err != redis.Nil {
			log.WithError(err).Warn("Failed to get data from Redis")
			return nil, fmt.Errorf("failed to get cached query: %w", err)
		}

		return nil, nil // No cached data found
	}

	if err := json.Unmarshal(data, &query); err != nil {
		log.WithError(err).Error("Failed to unmarshal cached query")
		return nil, fmt.Errorf("failed to unmarshal cached query: %w", err)
	}

	log.Debug("Retrieved query from user-offer cache")
	return &query, nil
}

// CacheQuery stores a query in the cache
func (cache userOfferCache) CacheQuery(ctx context.Context, query domain.Query) error {
	data, err := json.Marshal(query)
	if err != nil {
		log.WithError(err).Error("Failed to marshal query for caching")
		return fmt.Errorf("failed to marshal query: %w", err)
	}

	key := cache.cacheKey(query)
	// Cache for 1 day
	if err := UserOfferCacheInstance.redisClient.Set(ctx, key, data, time.Hour*24).Err(); err != nil {
		log.WithError(err).Error("Failed to store query in Redis")
		return fmt.Errorf("failed to store query in Redis: %w", err)
	}

	log.Debugf("Stored query in user-offer cache with key: %s", key)
	return nil
}
