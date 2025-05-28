package domain

import (
	"crypto/sha256"
	"encoding/base64"
)

type Query struct {
	Offers    map[string]Offer   `json:"offers,omitempty"`
	Address   Address   `json:"address"`
	Timestamp int64 `json:"timestamp"`
	SessionID string    `json:"sessionId"`

	// helper fields

	// hash over address
	HelperAddressHash string `json:"addressHash"`
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
