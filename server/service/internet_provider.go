package service

import (
	"context"
	"server/domain"
	"server/utils"
)

type InternetProviderAPI interface {
	GetOffersStream(ctx context.Context, address domain.Address, offersChannel *utils.PubSubChannel[domain.Offer], errChannel chan<- error)
	GetProviderName() string
}
