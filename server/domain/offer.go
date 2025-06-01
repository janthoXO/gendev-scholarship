package domain

import (
	"fmt"
	"server/utils"
	"strings"
)

type Offer struct {
	// product details

	Provider                     string            `json:"provider"`
	ProductID                    int               `json:"productId,omitzero"`
	ProductName                  string            `json:"productName"`
	Speed                        int               `json:"speed"`
	ContractDurationInMonths     int               `json:"contractDurationInMonths"`
	ConnectionType               ConnectionType    `json:"connectionType"`
	Tv                           string            `json:"tv,omitzero"`
	LimitInGb                    int               `json:"limitInGb,omitzero"`
	MaxAgePerson                 int               `json:"maxAgePerson,omitzero"`
	MonthlyCostInCent            int               `json:"monthlyCostInCent"`
	MonthlyCostInCentWithVoucher int               `json:"monthlyCostInCentWithVoucher,omitzero"`
	AfterTwoYearsMonthlyCost     int               `json:"afterTwoYearsMonthlyCost,omitzero"`
	InstallationService          bool              `json:"installationService"`
	VoucherDetails               VoucherDetails    `json:"voucherDetails,omitzero"`
	ExtraProperties              map[string]string `json:"extraProperties,omitzero"`

	// helper fields

	// hash over product details
	HelperOfferHash     string `json:"offerHash"`
	HelperIsPreliminary bool   `json:"isPreliminary"`
}

func (o *Offer) GenerateHash() {
	o.HelperOfferHash = utils.HashURLEncoded(fmt.Appendf(nil, "%s%d%s%d%d%s%s%d%d%d%d%d%t%s%v", o.Provider, o.ProductID, o.ProductName,
		o.Speed, o.ContractDurationInMonths, o.ConnectionType, o.Tv, o.LimitInGb, o.MaxAgePerson,
		o.MonthlyCostInCent, o.AfterTwoYearsMonthlyCost, o.MonthlyCostInCentWithVoucher, o.InstallationService,
		o.VoucherDetails.GetHash(), o.ExtraProperties))
}

// ConnectionType represents the type of internet connection
type ConnectionType string

// ConnectionType enum values
const (
	DSL    ConnectionType = "DSL"
	CABLE  ConnectionType = "CABLE"
	FIBER  ConnectionType = "FIBER"
	MOBILE ConnectionType = "MOBILE"
)

func (c ConnectionType) String() string {
	return string(c)
}

func FromStringToConnectionType(s string) ConnectionType {
	s = strings.ToUpper(s) // Normalize to uppercase for comparison
	switch s {
	case "DSL":
		return DSL
	case "CABLE":
		return CABLE
	case "FIBER":
		return FIBER
	case "MOBILE":
		return MOBILE
	default:
		return ConnectionType(s) // Return as is if not recognized
	}
}

type VoucherType string

const (
	ABSOLUTE   VoucherType = "ABSOLUTE"
	PERCENTAGE VoucherType = "PERCENTAGE"
)

type VoucherDetails struct {
	Type        VoucherType `json:"voucherType"`
	Value       int         `json:"voucherValue"`
	Description string      `json:"voucherDescription,omitempty"`
}

func (v *VoucherDetails) GetHash() string {
	return string(utils.Hash(fmt.Appendf(nil, "%s%d", v.Type, v.Value)))
}
