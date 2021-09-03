package main

import (
	"fmt"
	"github.com/go-kit/log"
	"sync"
)

// exchangeRate actor to manage exchange rate state of an individual currency
type exchangeRate struct {
}

type ProxyApi struct {
	// coinbase API to interact with Coinbase REST endpoints
	coinbase *CoinbaseApi

	// rates maps a currency code to a map of currencies to rates
	rates map[Currency]Rates

	// logger for logging
	logger log.Logger

	lock sync.RWMutex
}

func NewProxyApi(coinbase *CoinbaseApi, logger log.Logger) *ProxyApi {
	return &ProxyApi{
		coinbase: coinbase,
		logger:   logger,
		rates:    map[Currency]Rates{},
	}
}

func (api *ProxyApi) refreshNow(currency Currency) error {
	rates, err := api.coinbase.ExchangeRates(currency)
	if err != nil {
		return fmt.Errorf("refresh now [%v]: %w", currency, err)
	}
	api.lock.Lock()
	defer api.lock.Unlock()
	api.rates[currency] = rates
	return nil
}

func (api *ProxyApi) Convert(amount Amount, from Currency, to Currency) (Amount, error) {
	api.lock.RLock()
	rates, ok := api.rates[from]
	api.lock.RUnlock()

	if !ok {
		// note... slight race condition in that multiple request for the same yet-unused currency
		// will result in multiple updates for the same currency.
		err := api.refreshNow(from)
		if err != nil {
			return 0, err
		}
	}
	api.lock.RLock()
	defer api.lock.RUnlock()
	rates = api.rates[from] // no need to check ok because we know that refreshNow, if called, succeeded.

	rate, ok := rates[to]
	if !ok {
		fmt.Println(rate)
		return 0, fmt.Errorf("unknown 'to' currency: %v", to)
	}
	return Amount(float64(rate) * float64(amount)), nil
}
