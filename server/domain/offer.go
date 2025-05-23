package domain

import (
	"crypto/sha256"
	"fmt"
)

type Offer struct {
	// product details
	Provider                     string            `json:"provider"`
	ProductID                    int               `json:"productId"`
	ProductName                  string            `json:"ProductName"`
	Speed                        int               `json:"speed"`
	ContractDurationInMonths     int               `json:"contractDurationInMonths"`
	ConnectionType               string            `json:"connectionType"`
	Tv                           string            `json:"tv"`
	LimitInGb                    int               `json:"limitInGb"`
	MaxAgePerson                 int               `json:"maxAgePerson"`
	MonthlyCostInCent            int               `json:"monthlyCostInCent"`
	AfterTwoYearsMonthlyCost     int               `json:"afterTwoYearsMonthlyCost"`
	MonthlyCostInCentWithVoucher int               `json:"monthlyCostInCentWithVoucher"`
	InstallationService          bool              `json:"installationService"`
	VoucherType                  string            `json:"voucherType"`
	VoucherValue                 int               `json:"voucherValue"`
	ExtraProperties              map[string]string `json:"extraProperties"`

	// helper fields
	// hash over product details
	HelperOfferHash         string `json:"offerHash"`
}

func (o *Offer) GetHash() string {
	h := sha256.New()
	h.Write(fmt.Appendf(nil, "%s%d%s%d%d%s%s%d%d%d%d%d%t%s%d%v", o.Provider, o.ProductID, o.ProductName,
		o.Speed, o.ContractDurationInMonths, o.ConnectionType, o.Tv, o.LimitInGb, o.MaxAgePerson,
		o.MonthlyCostInCent, o.AfterTwoYearsMonthlyCost, o.MonthlyCostInCentWithVoucher, o.InstallationService,
		o.VoucherType, o.VoucherValue, o.ExtraProperties))
	return string(h.Sum(nil))
}
