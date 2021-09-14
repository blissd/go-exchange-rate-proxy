package coinbase

import (
	"context"
	"encoding/json"
	"fmt"
	"go-exchange-rate-proxy"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

const ApiUrlBase = "https://api.coinbase.com/v2"

// Service wraps the coinbase REST API
type Service interface {
	ExchangeRates(ctx context.Context, currency proxy.Currency) (proxy.Rates, error)
}

// service coinbase API
type service struct {
	// url base API url
	url string

	// client for HTTP requests
	client http.Client
}

// NewService constructs a valid coinbase Service.
func NewService() Service {
	return &service{
		url: ApiUrlBase,
		client: http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// ExchangeRates loads the current exchanges for a given currency.
// Service rates change every minute.
func (s *service) ExchangeRates(ctx context.Context, currency proxy.Currency) (proxy.Rates, error) {
	type Response struct {
		Data struct {
			Currency string
			Rates    map[string]string // maps currency codes to rates
		}
	}

	url := fmt.Sprintf("%v/exchange-rates?currency=%v", s.url, currency)

	request, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("building http request: %w", err)
	}
	httpResponse, err := s.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("http get: %w", err)
	}
	defer httpResponse.Body.Close()

	var response Response
	bytes, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		return nil, fmt.Errorf("reading json: %w", err)
	}

	err = json.Unmarshal(bytes, &response)
	if err != nil {
		return nil, fmt.Errorf("decoding json: %w", err)
	}

	rates := proxy.Rates{}
	for k, v := range response.Data.Rates {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, fmt.Errorf("bad rate value: %w", err)
		}
		rates[proxy.Currency(k)] = proxy.Rate(f)
	}

	return rates, nil
}
