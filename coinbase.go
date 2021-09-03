package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-kit/log"
	"io/ioutil"
	"net/http"
	"strconv"
)

const CoinbaseApiUrlBase = "https://api.coinbase.com/v2"

type CoinbaseApi struct {
	// url base API url
	url string

	logger log.Logger
}

func (api CoinbaseApi) ExchangeRates(currency Currency) (ExchangeRates, error) {
	type Response struct {
		Data struct {
			Currency string
			Rates    map[string]string // maps currency codes to rates
		}
	}

	url := fmt.Sprintf("%v/exchange-rates?currency=%v", api.url, currency)
	api.logger.Log("url", url)
	httpResponse, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("coinbase api: %w", err)
	}
	defer httpResponse.Body.Close()

	var response Response
	bytes, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		return nil, fmt.Errorf("reading json: %w", err)
	}

	api.logger.Log("json", string(bytes))

	err = json.Unmarshal(bytes, &response)
	if err != nil {
		return nil, fmt.Errorf("decoding json: %w", err)
	}

	rates := ExchangeRates{}
	for k, v := range response.Data.Rates {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, fmt.Errorf("bad rate value: %w", err)
		}
		rates[Currency(k)] = Rate(f)
	}

	return rates, nil
}
