package controller

import "server/domain"

type PingPerfectApi struct{}

func (api *PingPerfectApi) GetOffers(address domain.Address) (offers []domain.Offer, err error) {
	return offers, err
}
