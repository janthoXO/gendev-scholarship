package domain

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

type Offer struct {
	// product details

	Provider                     string            `json:"provider"`
	ProductID                    int               `json:"productId,omitzero"`
	ProductName                  string            `json:"ProductName"`
	Speed                        int               `json:"speed"`
	ContractDurationInMonths     int               `json:"contractDurationInMonths"`
	ConnectionType               string            `json:"connectionType"`
	Tv                           string            `json:"tv,omitzero"`
	LimitInGb                    int               `json:"limitInGb,omitzero"`
	MaxAgePerson                 int               `json:"maxAgePerson,omitzero"`
	MonthlyCostInCent            int               `json:"monthlyCostInCent"`
	AfterTwoYearsMonthlyCost     int               `json:"afterTwoYearsMonthlyCost,omitzero"`
	MonthlyCostInCentWithVoucher int               `json:"monthlyCostInCentWithVoucher,omitzero"`
	InstallationService          bool              `json:"installationService"`
	VoucherType                  string            `json:"voucherType,omitzero"`
	VoucherValue                 int               `json:"voucherValue,omitzero"`
	ExtraProperties              map[string]string `json:"extraProperties,omitzero"`

	// helper fields

	// hash over product details
	HelperOfferHash     string `json:"offerHash"`
	HelperIsPreliminary bool   `json:"isPreliminary"`
}

func (o *Offer) GenerateHash() {
	h := sha256.New()
	h.Write(fmt.Appendf(nil, "%s%d%s%d%d%s%s%d%d%d%d%d%t%s%d%v", o.Provider, o.ProductID, o.ProductName,
		o.Speed, o.ContractDurationInMonths, o.ConnectionType, o.Tv, o.LimitInGb, o.MaxAgePerson,
		o.MonthlyCostInCent, o.AfterTwoYearsMonthlyCost, o.MonthlyCostInCentWithVoucher, o.InstallationService,
		o.VoucherType, o.VoucherValue, o.ExtraProperties))
	o.HelperOfferHash = base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}
