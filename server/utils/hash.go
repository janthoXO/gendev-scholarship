package utils

import (
	"crypto/sha256"
	"encoding/base64"
)

func Hash(s []byte) []byte {
	h := sha256.New()
	h.Write(s)
	return h.Sum(nil)
}

func HashURLEncoded(s []byte) string {
	return base64.RawURLEncoding.EncodeToString(Hash(s))
}
