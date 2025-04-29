package controller

import "server/domain"

type InternetProviderAPI interface {
	GetOffers(domain.Address) ([]domain.Offer, error)
	AcceptOffer(offerID string) (string, error)
}