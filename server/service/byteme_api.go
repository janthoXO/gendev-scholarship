package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"server/domain"
	"server/utils"
)

type ByteMeApi struct{}

func (api *ByteMeApi) GetOffers(ctx context.Context, address domain.Address) (offers []domain.Offer, err error) {
	// Construct the API endpoint URL
	baseURL := "https://byteme.gendev7.check24.fun/app/api/products/data"
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	q := u.Query()
	q.Add("street", address.Street)
	q.Add("houseNumber", address.HouseNumber)
	q.Add("city", address.City)
	q.Add("plz", address.ZipCode)
	u.RawQuery = q.Encode()
	// Send the GET request with X-API-Key header
	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("X-Api-Key", utils.Cfg.ByteMe.ApiKey)

	// Send the request using the default HTTP client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned non-OK status: %d", resp.StatusCode)
	}

	// Read the CSV response
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Convert CSV to a slice of maps
	csvMaps, err := utils.CSVToMap(bodyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV data: %w", err)
	}

	// Convert maps to domain.Offer objects
	for _, item := range csvMaps {
		offer := api.mapToOffer(item)
		offers = append(offers, offer)
	}

	return offers, nil
}

// mapToOffer converts a map with properly typed values to a domain.Offer object
func (api *ByteMeApi) mapToOffer(item map[string]interface{}) (offer domain.Offer) {
	// Map product info fields
	if productId, ok := item["productId"].(int); ok {
		offer.ProductInfo.ProductID = productId
	}

	// Set provider name inside ProductInfo
	if providerName, ok := item["providerName"].(string); ok {
		offer.ProductInfo.ProviderName = providerName
	}

	if speed, ok := item["speed"].(int); ok {
		offer.ProductInfo.Speed = speed
	}

	if duration, ok := item["durationInMonths"].(int); ok {
		offer.ProductInfo.ContractDurationInMonths = duration
	}

	if limit, ok := item["limitFrom"].(int); ok {
		offer.ProductInfo.LimitFrom = limit
	}

	if age, ok := item["maxAge"].(int); ok {
		offer.ProductInfo.MaxAge = age
	}

	// Map string fields
	if connectionType, ok := item["connectionType"].(string); ok {
		offer.ProductInfo.ConnectionType = connectionType
	}

	if tv, ok := item["tv"].(string); ok {
		offer.ProductInfo.Tv = tv
	}

	// Map pricing details
	if cost, ok := item["monthlyCostInCent"].(int); ok {
		offer.PricingDetails.MonthlyCostInCent = cost
	}

	// Map pricing fields
	if afterTwoYearsCost, ok := item["afterTwoYearsMonthlyCost"].(int); ok {
		offer.PricingDetails.AfterTwoYearsMonthlyCost = afterTwoYearsCost
	}

	if service, ok := item["installationService"].(bool); ok {
		offer.PricingDetails.InstallationService = service
	}

	if voucherType, ok := item["voucherType"].(string); ok {
		offer.PricingDetails.VoucherType = voucherType
	}

	if voucherValue, ok := item["voucherValue"].(int); ok {
		offer.PricingDetails.VoucherValue = voucherValue
	}

	return offer
}

func (api *ByteMeApi) AcceptOffer(offerID string) (string, error) {
	// Not implemented for this challenge
	return "", nil
}

func (api *ByteMeApi) GetProviderName() string {
	return "ByteMe"
}
