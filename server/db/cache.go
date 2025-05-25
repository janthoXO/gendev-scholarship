package db

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"server/domain"
	"server/utils"
	"time"

	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
)

var (
	// RedisClient is the global Redis client instance
	OffersRedisClient *redis.Client
)

// Initialize the Redis client
func InitRedisClient() {
	OffersRedisClient = redis.NewClient(&redis.Options{
		Addr:     utils.Cfg.OfferCache.Url, // Redis server address
		Password: utils.Cfg.OfferCache.Password,               // No password set
		DB:       0,                // Use default DB
	})

	// Test connection
	ctx := context.Background()
	if err := OffersRedisClient.Ping(ctx).Err(); err != nil {
		log.WithError(err).Warn("Failed to connect to Redis")
	} else {
		log.Info("Connected to Redis successfully")
	}
}

// CacheKey generates a cache key from an address
func CacheKey(address domain.Address) string {
	h := sha256.New()
	// generate a unique key based on the address
	stringToHash := address.Street + address.HouseNumber + address.ZipCode + address.City
	
	h.Write([]byte(stringToHash))
	return string(h.Sum(nil))
}

// GetCachedOffers retrieves offers for an address from the cache
func GetCachedOffers(ctx context.Context, address domain.Address) ([]domain.Offer) {
	if OffersRedisClient == nil {
		return nil
	}

	key := CacheKey(address)
	data, err := OffersRedisClient.Get(ctx, key).Bytes()
	if err != nil {
		if err != redis.Nil {
			log.WithError(err).Warn("Failed to get data from Redis")
		}
		return nil
	}

	var offers []domain.Offer
	if err := json.Unmarshal(data, &offers); err != nil {
		log.WithError(err).Error("Failed to unmarshal cached offers")
		return nil
	}

	log.WithField("count", len(offers)).Debug("Retrieved offers from cache")
	return offers
}

// CacheOffers stores offers for an address in the cache
func CacheOffers(ctx context.Context, address domain.Address, offers []domain.Offer) {
	if OffersRedisClient == nil || len(offers) == 0 {
		return
	}

	data, err := json.Marshal(offers)
	if err != nil {
		log.WithError(err).Error("Failed to marshal offers for caching")
		return
	}

	key := CacheKey(address)
	// Cache for 1 day
	if err := OffersRedisClient.Set(ctx, key, data, time.Hour * 24).Err(); err != nil {
		log.WithError(err).Error("Failed to store offers in Redis")
	} else {
		log.WithField("count", len(offers)).Debug("Cached offers successfully")
	}
}
