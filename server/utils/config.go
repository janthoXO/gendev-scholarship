package utils

import (
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

type Configuration struct {
	Database struct {
		Url      string `env:"SHARE_DB_URL,notEmpty"`
		Name     string `env:"SHARE_DB_NAME,notEmpty"`
		User     string `env:"SHARE_DB_USER,notEmpty"`
		Password string `env:"SHARE_DB_PASSWORD,notEmpty"`
	}
	OfferCache struct {
		Url      string `env:"OFFER_CACHE_URL,notEmpty"`
		Password string `env:"OFFER_CACHE_PASSWORD,notEmpty"`
		TTL      int64  `env:"OFFER_CACHE_TTL_SEC" envDefault:"300"` // 5 minutes
	}
	UserOfferCache struct {
		Url      string `env:"USER_OFFER_CACHE_URL,notEmpty"`
		Password string `env:"USER_OFFER_CACHE_PASSWORD,notEmpty"`
		TTL      int64  `env:"USER_OFFER_CACHE_TTL_SEC" envDefault:"86400"` // 24 hours
	}
	Server struct {
		Port                uint   `env:"SERVER_PORT" envDefault:"8080"`
		FreshnessWindowSec  int64  `env:"FRESHNESS_WINDOW_SEC" envDefault:"5"`
		RetryFrequencyMilli []uint `env:"RETRY_FREQUENCY_MILLI" envDefault:"1000,2000,3000"`
		ApiTimeoutSec       uint   `env:"API_TIMEOUT_SEC" envDefault:"30"`
	}
	VerbynDich struct {
		ApiKey string `env:"VERBYNDICH_API_KEY,notEmpty"`
	}
	ServusSpeed struct {
		Username string `env:"SERVUSSPEED_USERNAME,notEmpty"`
		Password string `env:"SERVUSSPEED_PASSWORD,notEmpty"`
	}
	PingPerfect struct {
		SignatureSecret string `env:"PINGPERFECT_SIGNATURE_SECRET,notEmpty"`
		ClientId        string `env:"PINGPERFECT_CLIENT_ID,notEmpty"`
	}
	WebWunder struct {
		ApiKey string `env:"WEBWUNDER_API_KEY,notEmpty"`
	}
	ByteMe struct {
		ApiKey string `env:"BYTEME_API_KEY,notEmpty"`
	}

	Debug bool `env:"DEBUG" envDefault:"false"`
}

var (
	Cfg Configuration
)

func LoadConfig() Configuration {
	err := godotenv.Load(".env")
	if err != nil {
		log.WithError(err).Warn("Error loading .env file")
	}

	err = env.Parse(&Cfg)
	if err != nil {
		log.WithError(err).Fatal("Error parsing environment variables")
	}

	if Cfg.Debug {
		log.SetLevel(log.DebugLevel)
		log.Warn("DEBUG MODE ENABLED")
	}

	return Cfg
}
