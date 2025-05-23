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
	// hash over address and timestamp
	HelperQueryHash string `json:"queryHash"`
}

func (q *Query) GetHash() string {
	h := sha256.New()
	h.Write([]byte(q.Address.Street + q.Address.HouseNumber + q.Address.ZipCode + q.Address.City + q.Timestamp.String()))
	return string(h.Sum(nil))
}