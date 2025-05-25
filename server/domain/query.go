package domain

import (
	"crypto/sha256"
	"encoding/base64"
	"time"
)

type Query struct {
	Offers    []Offer   `json:"offers,omitempty"`
	Address   Address   `json:"address"`
	Timestamp time.Time `json:"timestamp"`
	SessionID string    `json:"sessionId"`

	// helper fields
	// hash over address, timestamp and offers
	HelperAddressHash string `json:"-"`
}

func (q *Query) GenerateAddressHash()  {
	q.HelperAddressHash = GetHashByAddress(q.Address)
}

func GetHashByAddress(address Address) string {
	h := sha256.New()
	// generate a unique key based on the address
	stringToHash := address.Street + address.HouseNumber + address.ZipCode + address.City

	h.Write([]byte(stringToHash))
    return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}
