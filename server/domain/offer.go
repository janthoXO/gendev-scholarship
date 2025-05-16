package domain

type Offer struct {
	ProductInfo struct {
		ProductID             int `json:"productId"`
		ProviderName string `json:"providerName"`
		Speed                 int    `json:"speed"`
		ContractDurationInMonths int    `json:"contractDurationInMonths"`
		ConnectionType        string `json:"connectionType"`
		Tv                    string `json:"tv"`
		LimitFrom             int    `json:"limitFrom"`
		MaxAge                int    `json:"maxAge"`
	} `json:"productInfo"`
	PricingDetails struct {
		MonthlyCostInCent     int    `json:"monthlyCostInCent"`
		AfterTwoYearsMonthlyCost int   `json:"afterTwoYearsMonthlyCost"`
		InstallationService    bool `json:"installationService"`
		VoucherType         string `json:"voucherType"`
		VoucherValue        int    `json:"voucherValue"`
	} `json:"pricingDetails"`
}
