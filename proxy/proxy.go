package proxy

import (
	"fmt"
	"github.com/go-kit/log"
	"go-exchange-rate-proxy/coinbase"
	"go-exchange-rate-proxy/domain"
	"sync"
)

type Api struct {
	// coinbase API to interact with Coinbase REST endpoints
	coinbase *coinbase.Api

	// rates maps a currency code to a map of currencies to rates
	rates map[domain.Currency]domain.Rates

	// lock synchronizes access to rates map
	lock sync.RWMutex

	// logger for logging
	logger log.Logger
}

func New(cb *coinbase.Api, logger log.Logger) *Api {
	return &Api{
		coinbase: cb,
		logger:   logger,
		rates:    map[domain.Currency]domain.Rates{},
	}
}

func (api *Api) Convert(amount domain.Amount, from domain.Currency, to domain.Currency) (domain.Amount, error) {
	api.lock.RLock()
	rates, ok := api.rates[from]
	api.lock.RUnlock()

	if !ok {
		// note... slight race condition in that multiple request for the same yet-unused currency
		// will result in multiple updates for the same currency.
		err := api.refresh(from)
		if err != nil {
			return 0, err
		}
	}
	api.lock.RLock()
	defer api.lock.RUnlock()
	rates = api.rates[from] // no need to check ok because we know that refresh, if called, succeeded.

	rate, ok := rates[to]
	if !ok {
		fmt.Println(rate)
		return 0, fmt.Errorf("unknown 'to' currency: %v", to)
	}
	return domain.Amount(float64(rate) * float64(amount)), nil
}

func (api *Api) refresh(currency domain.Currency) error {
	rates, err := api.coinbase.ExchangeRates(currency)
	if err != nil {
		return fmt.Errorf("refresh now [%v]: %w", currency, err)
	}
	api.lock.Lock()
	defer api.lock.Unlock()
	api.rates[currency] = rates
	return nil
}
