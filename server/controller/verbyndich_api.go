package controller

import "server/domain"

type VerbyndichAPI struct{}

func (api *VerbyndichAPI) GetOffers(address domain.Address) (offers []domain.Offer, err error) {
	return offers, err
}
