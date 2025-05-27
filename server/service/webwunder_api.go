package service

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"server/domain"
	"server/utils"
)

type WebWunderApi struct{}

// WebWunderSoapEnvelope represents the SOAP envelope for the request
type WebWunderSoapEnvelope struct {
	XMLName xml.Name          `xml:"soapenv:Envelope"`
	SoapNS  string            `xml:"xmlns:soapenv,attr"`
	GsNS    string            `xml:"xmlns:gs,attr"`
	Header  string            `xml:"soapenv:Header"`
	Body    WebWunderSoapBody `xml:"soapenv:Body"`
}

// WebWunderSoapBody represents the SOAP body for the request
type WebWunderSoapBody struct {
	XMLName                 xml.Name             `xml:"soapenv:Body"`
	LegacyGetInternetOffers WebWunderSoapRequest `xml:"gs:legacyGetInternetOffers"`
}

// WebWunderSoapRequest represents the request payload
type WebWunderSoapRequest struct {
	XMLName xml.Name           `xml:"gs:legacyGetInternetOffers"`
	Input   WebWunderSoapInput `xml:"gs:input"`
}

// WebWunderSoapInput represents the input parameters as per WSDL spec
type WebWunderSoapInput struct {
	XMLName        xml.Name             `xml:"gs:input"`
	Installation   bool                 `xml:"gs:installation"`
	ConnectionEnum string               `xml:"gs:connectionEnum"`
	Address        WebWunderSoapAddress `xml:"gs:address"`
}

// WebWunderSoapAddress represents the address parameters as per WSDL spec
type WebWunderSoapAddress struct {
	XMLName     xml.Name `xml:"gs:address"`
	Street      string   `xml:"gs:street"`
	HouseNumber string   `xml:"gs:houseNumber"`
	City        string   `xml:"gs:city"`
	PLZ         string   `xml:"gs:plz"`
	CountryCode string   `xml:"gs:countryCode"`
}

// WebWunderSoapResponse represents the SOAP response structure
type WebWunderSoapResponse struct {
	XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"`
	Body    struct {
		XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Body"`
		Output  struct {
			XMLName  xml.Name               `xml:"Output"`
			Products []WebWunderSoapProduct `xml:"products"`
		} `xml:"Output"`
	} `xml:"Body"`
}

// WebWunderSoapProduct represents a product in the SOAP response as per WSDL spec
type WebWunderSoapProduct struct {
	XMLName      xml.Name                  `xml:"products"`
	ProductID    int                       `xml:"productId"`
	ProviderName string                    `xml:"providerName"`
	ProductInfo  *WebWunderSoapProductInfo `xml:"productInfo,omitempty"`
}

// WebWunderSoapProductInfo represents the product info as per WSDL spec
type WebWunderSoapProductInfo struct {
	XMLName                        xml.Name              `xml:"productInfo"`
	Speed                          int                   `xml:"speed"`
	MonthlyCostInCent              int                   `xml:"monthlyCostInCent"`
	MonthlyCostInCentFrom25thMonth int                   `xml:"monthlyCostInCentFrom25thMonth"`
	Voucher                        *WebWunderSoapVoucher `xml:"voucher,omitempty"`
	ContractDurationInMonths       int                   `xml:"contractDurationInMonths"`
	ConnectionType                 string                `xml:"connectionType"`
}

// WebWunderSoapVoucher represents the voucher in the response
type WebWunderSoapVoucher struct {
	XMLName           xml.Name `xml:"voucher"`
	PercentageVoucher *struct {
		XMLName           xml.Name `xml:"percentageVoucher"`
		Percentage        int      `xml:"percentage"`
		MaxDiscountInCent int      `xml:"maxDiscountInCent"`
	} `xml:"percentageVoucher,omitempty"`
	AbsoluteVoucher *struct {
		XMLName             xml.Name `xml:"absoluteVoucher"`
		DiscountInCent      int      `xml:"discountInCent"`
		MinOrderValueInCent int      `xml:"minOrderValueInCent"`
	} `xml:"absoluteVoucher,omitempty"`
}

const (
	// connection types
	ConnectionTypeDSL    = "DSL"
	ConnectionTypeFiber  = "FIBER"
	ConnectionTypeCable  = "CABLE"
	ConnectionTypeMobile = "MOBILE"
)

func (api *WebWunderApi) GetOffersStream(ctx context.Context, address domain.Address, offersChannel chan<- domain.Offer, errChannel chan<- error) {
	// TODO if no connection type specified, query all in parallel

	// TODO if installation not specified, try both per connection type

	// Create SOAP request envelope
	soapEnvelope := WebWunderSoapEnvelope{
		SoapNS: "http://schemas.xmlsoap.org/soap/envelope/",
		GsNS:   "http://webwunder.gendev7.check24.fun/offerservice",
		Header: "",
		Body: WebWunderSoapBody{
			LegacyGetInternetOffers: WebWunderSoapRequest{
				Input: WebWunderSoapInput{
					Installation:   false,
					ConnectionEnum: "DSL",
					Address: WebWunderSoapAddress{
						Street:      address.Street,
						HouseNumber: address.HouseNumber,
						City:        address.City,
						PLZ:         address.ZipCode,
						CountryCode: "DE",
					},
				},
			},
		},
	}

	// Marshal the request to XML
	requestXML, err := xml.MarshalIndent(soapEnvelope, "", "  ")
	if err != nil {
		select {
		case <-ctx.Done():
			return
		case errChannel <- fmt.Errorf("%s: failed to marshal SOAP request: %w", api.GetProviderName(), err):
		}
		return
	}

	// Create XML declaration and prepend to the request
	xmlHeader := []byte(`<?xml version="1.0" encoding="UTF-8"?>`)
	requestXML = append(xmlHeader, requestXML...)

	// Send the request
	body, err := utils.RetryWrapper(ctx, func() ([]byte, error) {
		// Create HTTP request with the SOAP payload and context
		req, err := http.NewRequestWithContext(ctx, "POST", "https://webwunder.gendev7.check24.fun:443/endpunkte/soap/ws", bytes.NewReader(requestXML))
		if err != nil {
			return nil, fmt.Errorf("%s: failed to create request: %w", api.GetProviderName(), err)
		}

		// Set necessary headers
		req.Header.Set("Content-Type", "text/xml; charset=utf-8")
		req.Header.Set("X-Api-Key", utils.Cfg.WebWunder.ApiKey)
		req.Header.Set("SOAPAction", "legacyGetInternetOffers")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		// Check the response status code
		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("%s: received non-200 response: %d with body %s", api.GetProviderName(), resp.StatusCode, bodyBytes)
		}

		return io.ReadAll(resp.Body)
	})
	if err != nil {
		select {
		case <-ctx.Done():
			return
		case errChannel <- err:
		}
		return
	}

	// Unmarshal the XML response
	var soapResponse WebWunderSoapResponse
	if err := xml.Unmarshal(body, &soapResponse); err != nil {
		select {
		case <-ctx.Done():
			return
		case errChannel <- fmt.Errorf("%s: failed to unmarshal SOAP response: %w", api.GetProviderName(), err):
		}
		return
	}

	// Convert the SOAP response products to domain offers
	for _, product := range soapResponse.Body.Output.Products {
		offer := api.soapProductToOffer(product)
		offer.Provider = api.GetProviderName()
		offer.HelperOfferHash = offer.GetHash()
		select {
		case <-ctx.Done():
			return
		case offersChannel <- offer:
		}
	}
}

// soapProductToOffer converts a WebWunder SOAP product to a domain.Offer
func (api *WebWunderApi) soapProductToOffer(product WebWunderSoapProduct) (offer domain.Offer) {
	// Map product info
	offer.ProductID = product.ProductID
	offer.ProductName = product.ProviderName

	// Initialize values to defaults
	offer.InstallationService = false // Default for WebWunder

	if product.ProductInfo != nil {
		offer.Speed = product.ProductInfo.Speed
		offer.ContractDurationInMonths = product.ProductInfo.ContractDurationInMonths
		offer.ConnectionType = product.ProductInfo.ConnectionType

		// Map pricing details
		offer.MonthlyCostInCent = product.ProductInfo.MonthlyCostInCent
		offer.AfterTwoYearsMonthlyCost = product.ProductInfo.MonthlyCostInCentFrom25thMonth

		// TODO process voucher differently
		// Process voucher if available
		if product.ProductInfo.Voucher != nil {
			if product.ProductInfo.Voucher.PercentageVoucher != nil {
				offer.VoucherType = "percentage"
				offer.VoucherValue = product.ProductInfo.Voucher.PercentageVoucher.Percentage
			} else if product.ProductInfo.Voucher.AbsoluteVoucher != nil {
				offer.VoucherType = "absolute"
				offer.VoucherValue = product.ProductInfo.Voucher.AbsoluteVoucher.DiscountInCent
			}
		}
	}

	return offer
}

func (api *WebWunderApi) GetProviderName() string {
	return "WebWunder"
}
