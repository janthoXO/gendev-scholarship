package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"server/domain"
	"server/utils"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

type VerbyndichAPI struct{}

type VerbyndichResponse struct {
	Product     string `json:"product"`
	Description string `json:"description"`
	Last        bool   `json:"last"`
	Valid       bool   `json:"valid"`
}

func (api *VerbyndichAPI) GetOffers(ctx context.Context, address domain.Address) (offers []domain.Offer, err error) {
	// Format the address as required: "street;house number;city;plz"
	addressStr := fmt.Sprintf("%s;%s;%s;%s",
		address.Street,
		address.HouseNumber,
		address.City,
		address.ZipCode)

	// We need to fetch all pages
	page := 0
	lastPage := false

	for !lastPage {
		// Build the URL with query parameters
		baseURL := "https://verbyndich.gendev7.check24.fun/check24/data"
		u, err := url.Parse(baseURL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse URL: %w", err)
		}

		q := u.Query()
		q.Add("apiKey", utils.Cfg.VerbynDich.ApiKey)
		q.Add("page", strconv.Itoa(page))
		u.RawQuery = q.Encode()

		// Create the request with context
		req, err := http.NewRequestWithContext(ctx, "POST", u.String(), strings.NewReader(addressStr))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		// Send the request
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to send request: %w", err)
		}
		defer resp.Body.Close()

		// Check the response status
		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("API returned non-OK status: %d, body: %s", resp.StatusCode, string(bodyBytes))
		}

		// Decode the response
		var response VerbyndichResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}

		// Check if this is the last page
		lastPage = response.Last
		offer := domain.Offer{}
		offer.ProductInfo.ProviderName = response.Product

		// Process the offer if it's valid
		if response.Valid {
			// Parse the description to extract offer details
			err = api.parseVerbyndichDescription(response.Description, &offer)
			if err == nil {
				offers = append(offers, offer)
			}
		}

		// Move to the next page
		page++

		// Check if context is done to break the loop
		select {
		case <-ctx.Done():
			return offers, ctx.Err()
		default:
			// Continue
		}
	}

	return offers, nil
}

// have to do it in multiple regex because i was not able to write a regex with multiple options after each other
var regexPatterns = []func(string, *domain.Offer) error{
	func(description string, offer *domain.Offer) error {
		regexPattern := regexp.MustCompile(strings.Join([]string{
			`(?s)`,
			`.*?`,
			`nur\s+(?P<Price>\d+)€\s+im\s+Monat`,
			`.*?`,
			`(?P<Type>DSL|Cable|Fiber)\-Verbindung`,
			`.*?`,
			`Geschwindigkeit\s+von\s+(?P<Speed>\d+)\s+Mbit\/s`,
			`.*?`,
			`Mindestvertragslaufzeit\s+(?P<MinContract>\d+)\s*Monate`,
			`.*?`,
			`Ab\s+dem\s+24\.\s+Monat\s+beträgt\s+der\s+monatliche\s+Preis\s+(?P<TwoYearsCost>\d+)€`,
			`.*`,
		}, ""))
		names := regexPattern.SubexpNames()
		if m := regexPattern.FindStringSubmatch(description); m != nil {
			for i, name := range names {
				if name == "" || m[i] == "" {
					continue
				}
				switch name {
				case "Price":
					costInEuro, _ := strconv.Atoi(m[i])
					offer.PricingDetails.MonthlyCostInCent = costInEuro * 100
					log.Debugf("Price found: %d\n", offer.PricingDetails.MonthlyCostInCent)
				case "Type":
					offer.ProductInfo.ConnectionType = m[i]
					log.Debugf("ConnectionType found: %s\n", offer.ProductInfo.ConnectionType)
				case "Speed":
					offer.ProductInfo.Speed, _ = strconv.Atoi(m[i])
					log.Debugf("Speed found: %d\n", offer.ProductInfo.Speed)
				case "MinContract":
					offer.ProductInfo.ContractDurationInMonths, _ = strconv.Atoi(m[i])
					log.Debugf("MinContract found: %d\n", offer.ProductInfo.ContractDurationInMonths)
				case "TwoYearsCost":
					costInEuro, _ := strconv.Atoi(m[i])
					offer.PricingDetails.AfterTwoYearsMonthlyCost = costInEuro * 100
					log.Debugf("After Two Years Price found: %d\n", offer.PricingDetails.AfterTwoYearsMonthlyCost)
				}
			}
		} else {
			// non optional matches so return error
			return fmt.Errorf("no match for offer %s with description %s\n", offer.ProductInfo.ProviderName, description)
		}

		return nil
	}, // non optional matches
	func(description string, offer *domain.Offer) error {
		regexPattern := regexp.MustCompile(`(?s).*?Fernsehsender\s+enthalten\s+([^.]+).*`)
		if matches := regexPattern.FindStringSubmatch(description); matches != nil && len(matches) > 1 {
			// matches[0] is the full match, matches[1] is the first capture group
			offer.ProductInfo.Tv = matches[1]
			log.Debugf("TV found: %s\n", offer.ProductInfo.Tv)
		} else {
			log.Debugf("no tv found for %s\n", offer.ProductInfo.ProviderName)
		}

		return nil
	}, //optional
	func(description string, offer *domain.Offer) error {
		regexPattern := regexp.MustCompile(`(?s).*?Ab\s+(\d+)GB\s+pro\s+Monat\s+wird\s+die\s+Geschwindigkeit\s+gedrosselt.*`)
		if matches := regexPattern.FindStringSubmatch(description); matches != nil && len(matches) > 1 {
			// matches[0] is the full match, matches[1] is the first capture group
			offer.ProductInfo.LimitFrom, _ = strconv.Atoi(matches[1])
			log.Debugf("throttle found: %d\n", offer.ProductInfo.LimitFrom)
		} else {
			log.Debugf("no throttle found for %s\n", offer.ProductInfo.ProviderName)
		}

		return nil
	}, //optional
	func(description string, offer *domain.Offer) error {
		regexPattern := regexp.MustCompile(`(?s).*?nur\s+für\s+Personen\s+unter\s+(\d+)\s+Jahren.*`)
		if matches := regexPattern.FindStringSubmatch(description); matches != nil && len(matches) > 1 {
			// matches[0] is the full match, matches[1] is the first capture group
			offer.ProductInfo.MaxAge, _ = strconv.Atoi(matches[1])
			log.Debugf("MaxAge found: %d\n", offer.ProductInfo.MaxAge)
		} else {
			log.Debugf("no max age found for %s\n", offer.ProductInfo.ProviderName)
		}

		return nil
	}, //optional
	func(description string, offer *domain.Offer) error {
		regexPattern := regexp.MustCompile(`(?s).*?Rabatt\s+von\s+(\d+)%.*?maximale\s+Rabatt\s+beträgt\s+?(\d+)€.*`)
		if matches := regexPattern.FindStringSubmatch(description); matches != nil && len(matches) > 1 {
			// matches[0] is the full match, matches[1] is the first capture group
			offer.PricingDetails.VoucherType = "Percentage"
			offer.PricingDetails.VoucherValue, _ = strconv.Atoi(matches[1])
			log.Debugf("voucher found: %d\n", offer.PricingDetails.VoucherValue)
		} else {
			log.Debugf("no voucher found for %s\n", offer.ProductInfo.ProviderName)
		}

		return nil
	}, //optional
}

func (api *VerbyndichAPI) parseVerbyndichDescription(description string, offer *domain.Offer) error {
	for _, patternFunc := range regexPatterns {
		var err error
		err = patternFunc(description, offer)
		if err != nil {
			return fmt.Errorf("Error parsing description: %v\n", err)
		}
	}

	return nil
}

func (api *VerbyndichAPI) AcceptOffer(offerID string) (string, error) {
	// Not implemented for this challenge
	return "", nil
}

func (api *VerbyndichAPI) GetProviderName() string {
	return "VerbynDich"
}
