package service

import (
	"context"
	"server/domain"
)

type InternetProviderAPI interface {
	GetOffersStream(ctx context.Context, address domain.Address, offersChannel chan<- domain.Offer, errChannel chan<- error)
	GetProviderName() string
}
