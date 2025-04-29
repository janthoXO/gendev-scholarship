package domain

type Address struct {
	Street  string `json:"street"`
	HouseNumber   string `json:"house-number"`
	City    string `json:"city"`
	ZipCode string `json:"zip-code"`
}
