package proxy

import (
	"fmt"
	"github.com/go-kit/log"
	"go-exchange-rate-proxy/coinbase"
	"go-exchange-rate-proxy/domain"
	"sync"
	"time"
)

// Api proxy service API
type Api struct {
	// lookup to lookup exchange rates. lookup must be concurrency-safe
	lookup LookupFunc

	// logger for logging
	logger log.Logger
}

// LookupFunc for looking up exchange rates for a currency.
// Implementations must be concurrency-safe when invoked.
// Returned rates must be safe for concurrent reads.
type LookupFunc func(currency domain.Currency) (domain.Rates, error)

// New constructs a valid Api
func New(lookup LookupFunc, logger log.Logger) *Api {
	return &Api{
		lookup: lookup,
		logger: logger,
	}
}

// Convert computes a conversion from one currency to another with the current exchange rate.
// As a side-effect the cache of exchange rates might be updated.
func (api *Api) Convert(amount domain.Amount, from domain.Currency, to domain.Currency) (*domain.Exchanged, error) {
	api.logger.Log("msg", "converting currency", "from", from, "to", to, "amount", amount)
	rates, err := api.lookup(from)
	if err != nil {
		return nil, fmt.Errorf("convert from [%v]: %w", from, err)
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

	api.logger.Log("msg", "converted currency",
		"from", from,
		"to", to,
		"amount", amount,
		"rate", rate,
		"converted_amount", result.Amount,
	)

	return &result, nil
}

// LookupWithApi look up exchange rates directly by coinbase.Api
func LookupWithApi(api *coinbase.Api) LookupFunc {
	return api.ExchangeRates
}

// LookupWithCache decorates another lookup to add caching and refreshing
func LookupWithCache(lookup LookupFunc, updateFrequency time.Duration, logger log.Logger) LookupFunc {
	cache := map[domain.Currency]domain.Rates{}
	lock := sync.RWMutex{}

	// refresh get rates from wrapped lookup and update cache
	// bool return is true if this is the first time the currency has been seen
	refreshNow := func(currency domain.Currency) (domain.Rates, bool, error) {
		rates, err := lookup(currency)
		if err != nil {
			return nil, false, fmt.Errorf("refresh [%v]: %w", currency, err)
		}
		lock.Lock()
		defer lock.Unlock()
		_, ok := cache[currency]
		cache[currency] = rates
		return rates, !ok, nil
	}

	// refreshPeriodically refreshes rate for one currency every updateFrequency.
	// Must be invoked from a go routine.
	refreshPeriodically := func(currency domain.Currency) {
		for {
			select {
			case <-time.After(updateFrequency):
				logger.Log("msg", "periodic refresh", "currency", currency)
				_, _, err := refreshNow(currency)
				if err != nil {
					// Don't return, just log and hope this is a transient error
					logger.Log("msg", "periodic refresh failed", "currency", currency, "error", err)
				}
			}
		}
	}

	return func(currency domain.Currency) (domain.Rates, error) {
		logger.Log("msg", "checking cache", "currency", currency)
		lock.RLock()
		rates, ok := cache[currency]
		lock.RUnlock()

		if !ok {
			logger.Log("msg", "seeding cache", "currency", currency)
			var err error // separate var err, so it is very clear that rates is being re-assigned below
			var firstTime bool

			// Note there is a race condition here in that multiple requests for a currency that isn't yet cached
			// will result in multiple concurrent attempts to refresh. This should be harmless, unless the underlying
			// coinbase API throttles the requests. We could avoid this by holding a lock while calling the
			// calling and waiting on the underlying coinbase API, but that is a blocking operation so I'd rather not.
			// To avoid running multiple go routines to periodically refresh the same currency, the refreshNow
			// function will inform of the first time the currency is cached.
			rates, firstTime, err = refreshNow(currency)
			if err != nil {
				return nil, fmt.Errorf("refreshing cache [%v]: %w", currency, err)
			}
			if firstTime {
				logger.Log("msg", "scheduling periodic refresh", "currency", currency)
				go refreshPeriodically(currency)
			}
		}

		return rates, nil
	}
}
