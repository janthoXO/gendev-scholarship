package controller

import "server/domain"

type WebWunderApi struct{}

func (api *WebWunderApi) GetOffers(address domain.Address) (offers []domain.Offer, err error) {
	return offers, err
}
