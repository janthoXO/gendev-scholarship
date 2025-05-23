package domain

import (
	"crypto/sha256"
	"time"
)

type Query struct {
	Offers []Offer `json:"offers"`
	Address Address `json:"address"`
	Timestamp time.Time `json:"timestamp"`

	// helper fields
	// hash over address, timestamp and offers
	HelperQueryHash string `json:"queryHash"`
	HelperNumberOffers int `json:"numberOfOffers"`
}

func (q *Query) GetHash() string {
	h := sha256.New()
	// generate a unique key based on the address and timestamp
	stringToHash := q.Address.Street + q.Address.HouseNumber + q.Address.ZipCode + q.Address.City + q.Timestamp.String()
	
	h.Write([]byte(stringToHash))
	return string(h.Sum(nil))
}
