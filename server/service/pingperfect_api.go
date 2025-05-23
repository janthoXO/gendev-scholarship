package service

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"server/domain"
	"server/utils"
	"strconv"
	"time"
)

type PingPerfectApi struct{}

type PingPerfectRequest struct {
	Street      string `json:"street"`
	PLZ         string `json:"plz"`
	HouseNumber string `json:"houseNumber"`
	City        string `json:"city"`
	WantsFiber  bool   `json:"wantsFiber"`
}

type PingPerfectProduct struct {
	ProviderName string `json:"providerName"`
	ProductInfo  struct {
		Speed                    int    `json:"speed"`
		ContractDurationInMonths int    `json:"contractDurationInMonths"`
		ConnectionType           string `json:"connectionType"`
		Tv                       string `json:"tv,omitempty"`
		LimitFrom                int    `json:"limitFrom,omitempty"`
		MaxAge                   int    `json:"maxAge,omitempty"`
	} `json:"productInfo,omitempty"`
	PricingDetails struct {
		MonthlyCostInCent   int    `json:"monthlyCostInCent"`
		InstallationService string `json:"installationService,omitempty"`
	} `json:"pricingDetails,omitempty"`
}

func (api *PingPerfectApi) GetOffersStream(ctx context.Context, address domain.Address, offersChannel chan<- domain.Offer, errChannel chan<- error) {
	// Create request payload
	requestData := PingPerfectRequest{
		Street:      address.Street,
		PLZ:         address.ZipCode,
		HouseNumber: address.HouseNumber,
		City:        address.City,
		WantsFiber:  false, // Set this based on user preference or default to true
	}

	// Convert request to JSON
	requestBody, err := json.Marshal(requestData)
	if err != nil {
		select {
		case <-ctx.Done():
			return
		case errChannel <- fmt.Errorf("%s: failed to marshal request data: %w", api.GetProviderName(), err):
		}
		return
	}

	// Generate timestamp and signature
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	signature := generatePingPerfectSignature(requestBody, timestamp, utils.Cfg.PingPerfect.SignatureSecret)

	// Create HTTP request with context
	req, err := http.NewRequestWithContext(ctx, "POST", "https://pingperfect.gendev7.check24.fun/internet/angebote/data", bytes.NewBuffer(requestBody))
	if err != nil {
		select {
		case <-ctx.Done():
			return
		case errChannel <- fmt.Errorf("%s: failed to create request: %w", api.GetProviderName(), err):
		}
		return
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Client-Id", utils.Cfg.PingPerfect.ClientId)
	req.Header.Set("X-Timestamp", timestamp)
	req.Header.Set("X-Signature", signature)

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		select {
		case <-ctx.Done():
			return
		case errChannel <- fmt.Errorf("%s: failed to send request: %w", api.GetProviderName(), err):
		}
		return
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		select {
		case <-ctx.Done():
			return
		case errChannel <- fmt.Errorf("%s: API returned non-OK status: %d, body: %s", api.GetProviderName(), resp.StatusCode, string(bodyBytes)):
		}
		return
	}

	// Parse response
	var products []PingPerfectProduct
	if err := json.NewDecoder(resp.Body).Decode(&products); err != nil {
		select {
		case <-ctx.Done():
			return
		case errChannel <- fmt.Errorf("%s: failed to decode response: %w", api.GetProviderName(), err):
		}
		return
	}

	// Convert to domain.Offer objects
	for _, product := range products {
		offer := api.productToOffer(product)
		offer.Provider = api.GetProviderName()
		offer.HelperOfferHash = offer.GetHash()
		select {
		case <-ctx.Done():
			return
		case offersChannel <- offer:
		}
	}
}

// productToOffer converts a PingPerfectProduct to a domain.Offer object
func (api *PingPerfectApi) productToOffer(product PingPerfectProduct) domain.Offer {
	offer := domain.Offer{}

	offer.ProductName = product.ProviderName
	offer.Speed = product.ProductInfo.Speed
	offer.ContractDurationInMonths = product.ProductInfo.ContractDurationInMonths
	offer.ConnectionType = product.ProductInfo.ConnectionType
	offer.Tv = product.ProductInfo.Tv
	offer.LimitInGb = product.ProductInfo.LimitFrom
	offer.MaxAgePerson = product.ProductInfo.MaxAge

	offer.MonthlyCostInCent = product.PricingDetails.MonthlyCostInCent
	offer.InstallationService = product.PricingDetails.InstallationService == "yes"

	return offer
}

// Generate HMAC-SHA256 signature for PingPerfect API
func generatePingPerfectSignature(requestBody []byte, timestamp, signatureSecret string) string {
	// Concatenate timestamp and request body with a colon separator
	dataToSign := timestamp + ":" + string(requestBody)

	// Create a new HMAC by defining the hash type and the key
	h := hmac.New(sha256.New, []byte(signatureSecret))

	// Write data to the HMAC
	h.Write([]byte(dataToSign))

	// Get result and encode as hexadecimal string
	return hex.EncodeToString(h.Sum(nil))
}

func (api *PingPerfectApi) AcceptOffer(offerID string) (string, error) {
	// Not implemented for this challenge
	return "", nil
}

func (api *PingPerfectApi) GetProviderName() string {
	return "PingPerfect"
}
