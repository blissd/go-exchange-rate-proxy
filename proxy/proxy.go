package proxy

import (
	"fmt"
	"github.com/go-kit/log"
	"go-exchange-rate-proxy/coinbase"
	"go-exchange-rate-proxy/domain"
	"sync"
)

// Api proxy service API
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

// New constructs a valid Api
func New(cb *coinbase.Api, logger log.Logger) *Api {
	return &Api{
		coinbase: cb,
		logger:   logger,
		rates:    map[domain.Currency]domain.Rates{},
	}
}

// Convert computes a conversion from one currency to another with the current exchange rate.
// As a side-effect the cache of exchange rates might be updated.
func (api *Api) Convert(amount domain.Amount, from domain.Currency, to domain.Currency) (*domain.Exchanged, error) {
	api.lock.RLock()
	rates, ok := api.rates[from]
	api.lock.RUnlock()

	if !ok {
		// note... slight race condition in that multiple requests for the same yet-unused currency
		// will result in multiple updates for the same currency, which is probably harmless.
		// We could avoid this with more locking, but I don't want to hold a lock for the duration of a HTTP
		// request to the coinbase API.
		var err error
		rates, err = api.refresh(from)
		if err != nil {
			return nil, err
		}
	}

	rate, ok := rates[to]
	if !ok {
		fmt.Println(rate)
		return nil, fmt.Errorf("unknown 'to' currency: %v", to)
	}

	result := domain.Exchanged{
		Rate:   rate,
		Amount: domain.Amount(float64(rate) * float64(amount)),
	}
	return &result, nil
}

// refresh exchange rates for a currency and return.
func (api *Api) refresh(currency domain.Currency) (domain.Rates, error) {
	api.logger.Log("msg", "refreshing", "currency", currency)
	rates, err := api.coinbase.ExchangeRates(currency)
	if err != nil {
		return nil, fmt.Errorf("refresh now [%v]: %w", currency, err)
	}
	api.lock.Lock()
	defer api.lock.Unlock()
	api.rates[currency] = rates
	return rates, nil
}
