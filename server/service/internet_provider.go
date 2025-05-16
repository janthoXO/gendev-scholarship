package service

import (
	"context"
	"server/domain"
)

type InternetProviderAPI interface {
	GetOffers(ctx context.Context, address domain.Address) ([]domain.Offer, error)
	AcceptOffer(offerID string) (string, error)
	GetProviderName() string
}
