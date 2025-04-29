package controller

import (
	"server/domain"
)

type ByteMeApi struct{}

func (api *ByteMeApi) GetOffers(address domain.Address) (offers []domain.Offer, err error) {
	return offers, err
}
