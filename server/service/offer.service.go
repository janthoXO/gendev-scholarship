package service

import (
	"context"
	"server/domain"
	"server/utils"
	"sync"
	"time"
)

type OfferServiceImpl struct{}

var providers = []InternetProviderAPI{
	&ByteMeApi{},
	&PingPerfectApi{},
	&ServusSpeedApi{},
	&VerbyndichAPI{},
	&WebWunderApi{},
}

func (service OfferServiceImpl) FetchOffersStream(ctx context.Context, address domain.Address, offersChannel chan<- domain.Offer, errChannel chan<- error) {
	// Create a parent context with the API timeout as a control mechanism
	// Using a separate context instead of wrapping the incoming one to avoid premature cancellation
	timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), time.Duration(utils.Cfg.Server.ApiTimeout)*time.Second)

	var wg sync.WaitGroup

	// Start goroutines for each provider
	for _, provider := range providers {
		wg.Add(1)
		go func(p InternetProviderAPI) {
			defer wg.Done()

			// Create a provider-specific context that gets canceled either by the timeout or the original context
			providerCtx, providerCancel := context.WithCancel(context.Background())
			defer providerCancel()

			// Monitor both contexts
			go func() {
				select {
				case <-ctx.Done():
					// The original request context was canceled (e.g., client disconnected)
					providerCancel()
				case <-timeoutCtx.Done():
					// The timeout occurred
					providerCancel()
				}
			}()

			// Call the streaming method for each provider
			p.GetOffersStream(providerCtx, address, offersChannel, errChannel)
		}(provider)
	}

	// Wait for all providers to complete in a separate goroutine
	go func() {
		wg.Wait()

		// Close channels when all providers are done
		close(offersChannel)
		close(errChannel)

		// Cleanup
		timeoutCancel()
	}()
}
