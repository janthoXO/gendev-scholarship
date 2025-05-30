package domain

type Address struct {
	Street  string `json:"street"`
	HouseNumber   string `json:"houseNumber"`
	City    string `json:"city"`
	ZipCode string `json:"zipCode"`
}
