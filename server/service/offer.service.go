package service

import (
	"context"
	"server/domain"
	"server/utils"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type OfferServiceImpl struct{}

var providers = []InternetProviderAPI{
	&ByteMeApi{},
	&PingPerfectApi{},
	&ServusSpeedApi{},
	&VerbyndichAPI{},
	&WebWunderApi{},
}

func (service OfferServiceImpl) FetchOffers(address domain.Address) (allOffers []domain.Offer, err error) {
	// Fetch offers from all providers concurrently
	var wg sync.WaitGroup
	offersChan := make(chan []domain.Offer, len(providers))
	errorsChan := make(chan error, len(providers))

	// Create a parent context that will be passed to all provider API calls
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(utils.Cfg.Server.ApiTimeout)*time.Second)
	defer cancel() // Ensure all child operations are canceled when this function returns

	for _, provider := range providers {
		wg.Add(1)
		go func(p InternetProviderAPI) {
			defer wg.Done()

			done := make(chan bool)
			var offers []domain.Offer
			var err error

			go func() {
				// Pass the parent context to GetOffers so it can be used in HTTP requests
				offers, err = p.GetOffers(ctx, address)
				done <- true
			}()

			select {
			case <-done:
				if err != nil {
					log.WithError(err).WithField("provider", p.GetProviderName()).Warn("Provider API error")
					errorsChan <- err
					return
				}
				offersChan <- offers
			case <-ctx.Done():
				log.WithField("provider", p.GetProviderName()).Warn("Provider API timeout")
				errorsChan <- ctx.Err()
			}
		}(provider)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(offersChan)
	close(errorsChan)

	// Collect all offers
	for offers := range offersChan {
		allOffers = append(allOffers, offers...)
	}

	return allOffers, nil
}
