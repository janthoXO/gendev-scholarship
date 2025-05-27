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

func (api *ByteMeApi) GetOffersStream(ctx context.Context, address domain.Address, offersChannel chan<- domain.Offer, errChannel chan<- error) {
	// Construct the API endpoint URL
	baseURL := "https://byteme.gendev7.check24.fun/app/api/products/data"
	u, err := url.Parse(baseURL)
	if err != nil {
		select {
		case <-ctx.Done():
			return
		case errChannel <- fmt.Errorf("%s: failed to parse URL: %w", api.GetProviderName(), err):
		}
		return
	}

	q := u.Query()
	q.Add("street", address.Street)
	q.Add("houseNumber", address.HouseNumber)
	q.Add("city", address.City)
	q.Add("plz", address.ZipCode)
	u.RawQuery = q.Encode()

	// Send the request using the default HTTP client
	bodyBytes, err := utils.RetryWrapper(ctx, func() ([]byte, error) {
		// Send the GET request with X-API-Key header
		req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to create request: %w", api.GetProviderName(), err)
		}
		req.Header.Set("X-Api-Key", utils.Cfg.ByteMe.ApiKey)

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

		// Read the CSV response
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

	// Convert CSV to a slice of maps
	csvMaps, err := utils.CSVToMap(bodyBytes)
	if err != nil {
		select {
		case <-ctx.Done():
			return
		case errChannel <- fmt.Errorf("%s: failed to parse CSV data: %w", api.GetProviderName(), err):
		}
		return
	}

	// Convert maps to domain.Offer objects
	for _, item := range csvMaps {
		offer := api.mapToOffer(item)
		offer.Provider = api.GetProviderName()
		offer.HelperOfferHash = offer.GetHash()
		select {
		case <-ctx.Done():
			return
		case offersChannel <- offer:
		}
	}
}

// mapToOffer converts a map with properly typed values to a domain.Offer object
func (api *ByteMeApi) mapToOffer(item map[string]interface{}) (offer domain.Offer) {
	// Map product info fields
	if productId, ok := item["productId"].(int); ok {
		offer.ProductID = productId
	}

	// Set provider name inside ProductInfo
	if providerName, ok := item["providerName"].(string); ok {
		offer.ProductName = providerName
	}

	if speed, ok := item["speed"].(int); ok {
		offer.Speed = speed
	}

	if duration, ok := item["durationInMonths"].(int); ok {
		offer.ContractDurationInMonths = duration
	}

	if limit, ok := item["limitFrom"].(int); ok {
		offer.LimitInGb = limit
	}

	if age, ok := item["maxAge"].(int); ok {
		offer.MaxAgePerson = age
	}

	// Map string fields
	if connectionType, ok := item["connectionType"].(string); ok {
		offer.ConnectionType = connectionType
	}

	if tv, ok := item["tv"].(string); ok {
		offer.Tv = tv
	}

	// Map pricing details
	if cost, ok := item["monthlyCostInCent"].(int); ok {
		offer.MonthlyCostInCent = cost
	}

	// Map pricing fields
	if afterTwoYearsCost, ok := item["afterTwoYearsMonthlyCost"].(int); ok {
		offer.AfterTwoYearsMonthlyCost = afterTwoYearsCost
	}

	if service, ok := item["installationService"].(bool); ok {
		offer.InstallationService = service
	}

	if voucherType, ok := item["voucherType"].(string); ok {
		offer.VoucherType = voucherType
	}

	if voucherValue, ok := item["voucherValue"].(int); ok {
		offer.VoucherValue = voucherValue
	}

	return offer
}

func (api *ByteMeApi) GetProviderName() string {
	return "ByteMe"
}
