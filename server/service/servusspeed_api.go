package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"server/domain"
	"server/utils"
	"sync"

	log "github.com/sirupsen/logrus"
)

type ServusSpeedApi struct{}

type ServusSpeedRequestAddress struct {
	Strasse      string `json:"strasse"`
	Hausnummer   string `json:"hausnummer"`
	Postleitzahl string `json:"postleitzahl"`
	Stadt        string `json:"stadt"`
	Land         string `json:"land"`
}

type ServusSpeedRequest struct {
	Address ServusSpeedRequestAddress `json:"address"`
}

type ServusSpeedProductsResponse struct {
	AvailableProducts []string `json:"availableProducts"`
}

type ServusSpeedProductResponse struct {
	ServusSpeedProduct struct {
		ProviderName string `json:"providerName"`
		ProductInfo  struct {
			Speed                    int    `json:"speed"`
			ContractDurationInMonths int    `json:"contractDurationInMonths"`
			ConnectionType           string `json:"connectionType"`
			Tv                       string `json:"tv,omitempty"`
			LimitFrom                int    `json:"limitFrom,omitempty"`
			MaxAge                   int    `json:"maxAge,omitempty"`
		} `json:"productInfo"`
		PricingDetails struct {
			MonthlyCostInCent   int  `json:"monthlyCostInCent"`
			InstallationService bool `json:"installationService"`
		} `json:"pricingDetails"`
		Discount int `json:"discount"`
	} `json:"servusSpeedProduct"`
}

func (api *ServusSpeedApi) GetOffersStream(ctx context.Context, address domain.Address, offersChannel chan<- domain.Offer, errChannel chan<- error) {
	// Step 1: Get the list of available product IDs
	productIDs, err := api.getAvailableProducts(ctx, address)
	if err != nil {
		select {
		case <-ctx.Done():
			return
		case errChannel <- err:
		}
		return
	}

	// Step 2: Get the details for each product ID in parallel
	var wg sync.WaitGroup

	for _, productID := range productIDs {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()

			product, err := api.getProductDetails(ctx, id, address)
			if err != nil {
				// Log the error but continue with other products
				log.WithError(err).WithField("productID", id).
					WithField("provider", "ServusSpeed").
					Warn("Failed to get product details")
				select {
				case <-ctx.Done():
					return
				case errChannel <- err:
				}
				return
			}

			// Convert to domain.Offer
			offer := api.convertToOffer(product)
			offer.Provider = api.GetProviderName()
			offer.HelperOfferHash = offer.GetHash()

			// Write directly to the passed channel
			select {
			case <-ctx.Done():
				return
			case offersChannel <- offer:
			}
		}(productID)
	}

	// Wait for all goroutines to complete
	wg.Wait()
}

func (api *ServusSpeedApi) getAvailableProducts(ctx context.Context, address domain.Address) ([]string, error) {
	// Create request body
	reqBody := ServusSpeedRequest{
		Address: ServusSpeedRequestAddress{
			Strasse:      address.Street,
			Hausnummer:   address.HouseNumber,
			Postleitzahl: address.ZipCode,
			Stadt:        address.City,
			Land:         "DE", // As per the README, only DE is supported
		},
	}

	// Convert to JSON
	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create HTTP request with context
	req, err := http.NewRequestWithContext(ctx, "POST", "https://servus-speed.gendev7.check24.fun/api/external/available-products", bytes.NewBuffer(reqJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add basic auth header
	auth := utils.Cfg.ServusSpeed.Username + ":" + utils.Cfg.ServusSpeed.Password
	basicAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
	req.Header.Set("Authorization", basicAuth)
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned non-OK status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse response
	var productsResp ServusSpeedProductsResponse
	if err := json.NewDecoder(resp.Body).Decode(&productsResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return productsResp.AvailableProducts, nil
}

func (api *ServusSpeedApi) getProductDetails(ctx context.Context, productID string, address domain.Address) (*ServusSpeedProductResponse, error) {
	// Create request body
	reqBody := ServusSpeedRequest{
		Address: ServusSpeedRequestAddress{
			Strasse:      address.Street,
			Hausnummer:   address.HouseNumber,
			Postleitzahl: address.ZipCode,
			Stadt:        address.City,
			Land:         "DE", // As per the README, only DE is supported
		},
	}

	// Convert to JSON
	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create HTTP request with context
	url := fmt.Sprintf("https://servus-speed.gendev7.check24.fun/api/external/product-details/%s", productID)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add basic auth header
	auth := utils.Cfg.ServusSpeed.Username + ":" + utils.Cfg.ServusSpeed.Password
	basicAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
	req.Header.Set("Authorization", basicAuth)
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned non-OK status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse response
	var productResp ServusSpeedProductResponse
	if err := json.NewDecoder(resp.Body).Decode(&productResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &productResp, nil
}

func (api *ServusSpeedApi) convertToOffer(product *ServusSpeedProductResponse) domain.Offer {
	offer := domain.Offer{}

	offer.ProductName = product.ServusSpeedProduct.ProviderName
	offer.Speed = product.ServusSpeedProduct.ProductInfo.Speed
	offer.ContractDurationInMonths = product.ServusSpeedProduct.ProductInfo.ContractDurationInMonths
	offer.ConnectionType = product.ServusSpeedProduct.ProductInfo.ConnectionType
	offer.Tv = product.ServusSpeedProduct.ProductInfo.Tv
	offer.LimitInGb = product.ServusSpeedProduct.ProductInfo.LimitFrom
	offer.MaxAgePerson = product.ServusSpeedProduct.ProductInfo.MaxAge

	// Handle the discount
	monthlyCostWithDiscount := product.ServusSpeedProduct.PricingDetails.MonthlyCostInCent - product.ServusSpeedProduct.Discount
	offer.MonthlyCostInCent = monthlyCostWithDiscount
	offer.InstallationService = product.ServusSpeedProduct.PricingDetails.InstallationService

	return offer
}

func (api *ServusSpeedApi) AcceptOffer(offerID string) (string, error) {
	// Not implemented for this challenge
	return "", nil
}

func (api *ServusSpeedApi) GetProviderName() string {
	return "ServusSpeed"
}
