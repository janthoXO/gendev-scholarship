package domain

type OfferStatus int

const (
	Preliminary OfferStatus = iota
	Valid
	Invalid
)