package domain

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
	HelperTimesNotAvailable int    `json:"timesNotAvailable"`
}
