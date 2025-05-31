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
	"sync"
	"sync/atomic"
)

type VerbyndichAPI struct{}

type VerbyndichResponse struct {
	Product     string `json:"product"`
	Description string `json:"description"`
	Last        bool   `json:"last"`
	Valid       bool   `json:"valid"`
}

func (api *VerbyndichAPI) GetOffersStream(ctx context.Context, address domain.Address, offersChannel *utils.PubSubChannel[domain.Offer], errChannel chan<- error) {
	// Format the address as required: "street;house number;city;plz"
	addressStr := fmt.Sprintf("%s;%s;%s;%s",
		address.Street,
		address.HouseNumber,
		address.City,
		address.ZipCode)

	// Worker pool setup
	const numWorkers = 5
	pageChannel := make(chan int, numWorkers*2) // Buffer to prevent blocking
	var lastPageFound int32                     // Atomic flag to signal when last page is found

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			api.worker(ctx, addressStr, pageChannel, &lastPageFound, offersChannel, errChannel)
		}()
	}

	// Page dispatcher - sends page numbers to workers
	go func() {
		defer close(pageChannel)
		page := 0
		for {
			// Check if we should stop dispatching
			if atomic.LoadInt32(&lastPageFound) == 1 {
				return
			}

			select {
			case <-ctx.Done():
				return
			case pageChannel <- page:
				page++
			}
		}
	}()

	// Wait for all workers to finish
	wg.Wait()
}

func (api *VerbyndichAPI) worker(ctx context.Context, addressStr string, pageChannel <-chan int, lastPageFound *int32, offersChannel *utils.PubSubChannel[domain.Offer], errChannel chan<- error) {
	for {
		select {
		case <-ctx.Done():
			return
		case page, ok := <-pageChannel:
			if !ok {
				return // Channel closed, no more work
			}

			// Process this page
			response, err := api.fetchPage(ctx, addressStr, page)
			if err != nil {
				select {
				case <-ctx.Done():
					return
				case errChannel <- err:
				}
				continue
			}

			// Check if this is the last page
			if response.Last {
				atomic.StoreInt32(lastPageFound, 1)
			}

			// Process the offer if it's valid
			if response.Valid {
				offer := domain.Offer{}
				offer.ProductName = response.Product

				if err := api.parseVerbyndichDescription(response.Description, &offer); err == nil {
					offer.Provider = api.GetProviderName()
					offer.HelperIsPreliminary = false
					offersChannel.Publish(offer)
				}
			}
		}
	}
}

func (api *VerbyndichAPI) fetchPage(ctx context.Context, addressStr string, page int) (*VerbyndichResponse, error) {
	// Build the URL with query parameters
	baseURL := "https://verbyndich.gendev7.check24.fun/check24/data"
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to parse URL: %w", api.GetProviderName(), err)
	}

	q := u.Query()
	q.Add("apiKey", utils.Cfg.VerbynDich.ApiKey)
	q.Add("page", strconv.Itoa(page))
	u.RawQuery = q.Encode()

	// Send the request
	return utils.RetryWrapper(ctx, func() (*VerbyndichResponse, error) {
		// Create the request with context
		req, err := http.NewRequestWithContext(ctx, "POST", u.String(), strings.NewReader(addressStr))
		if err != nil {
			return nil, fmt.Errorf("%s: failed to create request: %w", api.GetProviderName(), err)
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		// Check the response status
		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("%s: received non-200 response: %d with body %s", api.GetProviderName(), resp.StatusCode, bodyBytes)
		}

		var response VerbyndichResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		return &response, err
	})
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
					offer.MonthlyCostInCent = costInEuro * 100
				case "Type":
					offer.ConnectionType = domain.FromStringToConnectionType(m[i])
				case "Speed":
					offer.Speed, _ = strconv.Atoi(m[i])
				case "MinContract":
					offer.ContractDurationInMonths, _ = strconv.Atoi(m[i])
				case "TwoYearsCost":
					costInEuro, _ := strconv.Atoi(m[i])
					offer.AfterTwoYearsMonthlyCost = costInEuro * 100
				}
			}
		} else {
			// non optional matches so return error
			return fmt.Errorf("no match for offer %s with description %s", offer.ProductName, description)
		}

		return nil
	}, // non optional matches
	func(description string, offer *domain.Offer) error {
		regexPattern := regexp.MustCompile(`(?s).*?Fernsehsender\s+enthalten\s+([^.]+).*`)
		if matches := regexPattern.FindStringSubmatch(description); len(matches) > 1 {
			// matches[0] is the full match, matches[1] is the first capture group
			offer.Tv = matches[1]
		}

		return nil
	}, //optional
	func(description string, offer *domain.Offer) error {
		regexPattern := regexp.MustCompile(`(?s).*?Ab\s+(\d+)GB\s+pro\s+Monat\s+wird\s+die\s+Geschwindigkeit\s+gedrosselt.*`)
		if matches := regexPattern.FindStringSubmatch(description); len(matches) > 1 {
			// matches[0] is the full match, matches[1] is the first capture group
			offer.LimitInGb, _ = strconv.Atoi(matches[1])
		}

		return nil
	}, //optional
	func(description string, offer *domain.Offer) error {
		regexPattern := regexp.MustCompile(`(?s).*?nur\s+für\s+Personen\s+unter\s+(\d+)\s+Jahren.*`)
		if matches := regexPattern.FindStringSubmatch(description); len(matches) > 1 {
			// matches[0] is the full match, matches[1] is the first capture group
			offer.MaxAgePerson, _ = strconv.Atoi(matches[1])
		}

		return nil
	}, //optional
    func(description string, offer *domain.Offer) error {
        regexPattern := regexp.MustCompile(`(?s).*?(Rabatt\s+von\s+(\d+)%.*?maximale\s+Rabatt\s+beträgt\s+?(\d+)€).*`)
        if matches := regexPattern.FindStringSubmatch(description); len(matches) > 2 {
            // matches[0] is the full string
            // matches[1] is just the discount description text
            // matches[2] is the percentage value
            // matches[3] is the maximum discount in euros
            voucherValuePerc, _ := strconv.Atoi(matches[2])
            maxVoucherValueEuro, _ := strconv.Atoi(matches[3])
    
            offer.VoucherDetails = domain.VoucherDetails{
                Type:  domain.PERCENTAGE,
                Value: voucherValuePerc,
                Description: matches[1],
            }
    
            discountOverContractDuration := min((offer.MonthlyCostInCent * offer.ContractDurationInMonths * voucherValuePerc) / 100, maxVoucherValueEuro * 100)
            offer.MonthlyCostInCentWithVoucher = offer.MonthlyCostInCent - discountOverContractDuration / offer.ContractDurationInMonths
        }
    
        return nil
    }, //optional
}

func (api *VerbyndichAPI) parseVerbyndichDescription(description string, offer *domain.Offer) error {
	for _, patternFunc := range regexPatterns {
		err := patternFunc(description, offer)
		if err != nil {
			return fmt.Errorf("error parsing description: %v", err)
		}
	}

	return nil
}

func (api *VerbyndichAPI) GetProviderName() string {
	return "VerbynDich"
}
