package controller

import "server/domain"

type ServusSpeedApi struct{}

func (api *ServusSpeedApi) GetOffers(address domain.Address) (offers []domain.Offer, err error) {
	return offers, err
}
