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

func (service OfferServiceImpl) FetchOffersStream(ctx context.Context, address domain.Address) (*utils.PubSubChannel[domain.Offer], <-chan error) {
	// Create a parent context with the API timeout as a control mechanism
	// We derive from the incoming context so that client disconnects are properly propagated
	timeoutCtx, timeoutCancel := context.WithTimeout(ctx, time.Duration(utils.Cfg.Server.ApiTimeoutSec)*time.Second)

	// Create a done channel to signal completion
	offersChannel := utils.NewPubSubChannel[domain.Offer]()
	errChannel := make(chan error)

	var wg sync.WaitGroup

	// Start goroutines for each provider
	for _, provider := range providers {
		wg.Add(1)
		go func(p InternetProviderAPI) {
			defer wg.Done()

			// Create a provider-specific context derived from the timeout context
			// This ensures proper propagation of cancellation
			providerCtx, providerCancel := context.WithCancel(timeoutCtx)
			defer providerCancel()

			// Also monitor the original context for client disconnects
			go func() {
				<-ctx.Done()
				// The original request context was canceled (e.g., client disconnected)
				providerCancel()
			}()

			// Call the streaming method for each provider
			p.GetOffersStream(providerCtx, address, offersChannel, errChannel)
		}(provider)
	}

	// Wait for all providers to complete in a separate goroutine
	go func() {
		wg.Wait()

		// Signal that all providers are done
		offersChannel.Close()
		close(errChannel)

		// Cleanup
		timeoutCancel()
	}()

	// Return the done channel so the caller can wait for completion
	return offersChannel, errChannel
}
