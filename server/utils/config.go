package utils

import (
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

type Configuration struct {
	Database struct {
		Host     string `env:"DATABASE_HOST,notEmpty"`
		Port     uint   `env:"DATABASE_PORT,notEmpty"`
		Name     string `env:"DATABASE_NAME,notEmpty"`
		User     string `env:"DATABASE_USER,notEmpty"`
		Password string `env:"DATABASE_PASSWORD,notEmpty"`
	}
	OfferCache struct {
		Url      string `env:"OFFER_CACHE_URL,notEmpty"`
		Password string `env:"OFFER_CACHE_PASSWORD,notEmpty"`
	}
	UserOfferCache struct {
		Url      string `env:"USER_OFFER_CACHE_URL,notEmpty"`
		Password string `env:"USER_OFFER_CACHE_PASSWORD,notEmpty"`
	}
	Server struct {
		Port       uint `env:"SERVER_PORT" envDefault:"8080"`
		ApiTimeout uint `env:"API_TIMEOUT" envDefault:"30"`
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
	err := godotenv.Load()
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
