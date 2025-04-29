package domain

type Offer struct {
	ProviderName string `json:"providerName"`
	ProductInfo struct {
		Speed                 int    `json:"speed"`
		ContractDurationInMonths int    `json:"contractDurationInMonths"`
		ConnectionType        string `json:"connectionType"`
		Tv                    string `json:"tv"`
		LimitFrom             int    `json:"limitFrom"`
		MaxAge                int    `json:"maxAge"`
	} `json:"productInfo"`
	PricingDetails struct {
		MonthlyCostInCent     int    `json:"monthlyCostInCent"`
		InstallationService    string `json:"installationService"`
	} `json:"pricingDetails"`
}
