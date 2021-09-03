package coinbase

import (
	"encoding/json"
	"fmt"
	"github.com/go-kit/log"
	"go-exchange-rate-proxy/domain"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

const ApiUrlBase = "https://api.coinbase.com/v2"

// Api coinbase API
type Api struct {
	// url base API url
	url string

	// logger for... logging
	logger log.Logger

	// client for HTTP requests
	client http.Client
}

// New constructs a valid coinbase Api.
func New(logger log.Logger) *Api {
	return &Api{
		url:    ApiUrlBase,
		logger: logger,
		client: http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// ExchangeRates loads the current exchanges for a given currency.
// Exchange rates change every minute.
func (api *Api) ExchangeRates(currency domain.Currency) (domain.Rates, error) {
	type Response struct {
		Data struct {
			Currency string
			Rates    map[string]string // maps currency codes to rates
		}
	}

	url := fmt.Sprintf("%v/exchange-rates?currency=%v", api.url, currency)
	httpResponse, err := api.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("coinbase api: %w", err)
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

	rates := domain.Rates{}
	for k, v := range response.Data.Rates {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, fmt.Errorf("bad rate value: %w", err)
		}
		rates[domain.Currency(k)] = domain.Rate(f)
	}

	return rates, nil
}
