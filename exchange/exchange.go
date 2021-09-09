package exchange

import (
	"context"
	"fmt"
	"github.com/go-kit/log"
	"go-exchange-rate-proxy"
	"go-exchange-rate-proxy/coinbase"
	"sync"
	"time"
)

// Api proxy exchange API
type Api struct {
	// lookup to lookup exchange rates. lookup must be concurrency-safe
	lookup LookupFunc

	// logger for logging
	logger log.Logger
}

// LookupFunc for looking up exchange rates for a currency.
// Implementations must be concurrency-safe when invoked.
// Returned rates must be safe for concurrent reads.
type LookupFunc func(ctx context.Context, currency proxy.Currency) (proxy.Rates, error)

// New constructs a valid Api
func New(lookup LookupFunc, logger log.Logger) *Api {
	return &Api{
		lookup: lookup,
		logger: logger,
	}
}

// Convert computes a conversion from one currency to another with the current exchange rate.
// As a side-effect the cache of exchange rates might be updated.
func (api *Api) Convert(ctx context.Context, amount proxy.Amount, from proxy.Currency, to proxy.Currency) (*proxy.Exchanged, error) {
	api.logger.Log("msg", "converting currency", "from", from, "to", to, "amount", amount)
	rates, err := api.lookup(ctx, from)
	if err != nil {
		return nil, fmt.Errorf("convert from [%v]: %w", from, err)
	}

	rate, ok := rates[to]
	if !ok {
		fmt.Println(rate)
		return nil, fmt.Errorf("unknown 'to' currency: %v", to)
	}

	result := proxy.Exchanged{
		Rate:   rate,
		Amount: proxy.Amount(float64(rate) * float64(amount)),
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

// LookupWithCache decorates another lookup function to add caching and refreshing
func LookupWithCache(next LookupFunc, updateFrequency time.Duration, logger log.Logger) LookupFunc {
	cache := &cache{
		cache:           map[proxy.Currency]proxy.Rates{},
		updateFrequency: updateFrequency,
		lock:            sync.RWMutex{},
		next:            next,
		logger:          logger,
	}

	return cache.lookup
}

// cache of exchange rates. The cache is concurrency safe and will periodically refresh cached values.
type cache struct {
	// cache the cache of rates
	cache map[proxy.Currency]proxy.Rates

	// updateFrequency how often to refresh cached values
	updateFrequency time.Duration

	// lock synchronizes access to cache to make it concurrency safe
	lock sync.RWMutex

	// next the LookupFunc being decorated with a cache
	next LookupFunc

	// logger to log
	logger log.Logger
}

// refreshNow refreshes a cached entry immediately
func (c *cache) refreshNow(ctx context.Context, currency proxy.Currency) (proxy.Rates, bool, error) {
	rates, err := c.next(ctx, currency)
	if err != nil {
		return nil, false, fmt.Errorf("refresh [%v]: %w", currency, err)
	}
	c.lock.Lock()
	defer c.lock.Unlock()
	_, ok := c.cache[currency]
	c.cache[currency] = rates
	return rates, !ok, nil
}

// refreshPeriodically refreshes a cached entry on a given schedule.
// This is expected to be called from a go-routine for each currency.
func (c *cache) refreshPeriodically(ctx context.Context, currency proxy.Currency) {
	for {
		select {
		case <-time.After(c.updateFrequency):
			c.logger.Log("msg", "periodic refresh", "currency", currency)
			_, _, err := c.refreshNow(ctx, currency)
			if err != nil {
				// Don't return, just log and hope this is a transient error
				c.logger.Log("msg", "periodic refresh failed", "currency", currency, "error", err)
			}
		case <-ctx.Done():
			c.logger.Log("msg", "shutting down periodic refresh", "currency", currency)
			c.uncache(currency)
			return
		}
	}
}

// uncache safely removes currency from cache
func (c *cache) uncache(currency proxy.Currency) {
	c.lock.Lock()
	defer c.lock.Unlock()
	delete(c.cache, currency)
}

// lookup exchange rates and cache results
func (c *cache) lookup(ctx context.Context, currency proxy.Currency) (proxy.Rates, error) {
	c.logger.Log("msg", "checking cache", "currency", currency)
	c.lock.RLock()
	rates, ok := c.cache[currency]
	c.lock.RUnlock()

	if !ok {
		c.logger.Log("msg", "seeding cache", "currency", currency)

		// Note there is a race condition here in that multiple requests for a currency that isn't yet cached
		// will result in multiple concurrent attempts to refresh. This should be harmless, unless the underlying
		// coinbase API throttles the requests. We could avoid this by holding a lock while calling the
		// calling and waiting on the underlying coinbase API, but that is a blocking operation so I'd rather not.
		// To avoid running multiple go routines to periodically refresh the same currency, the refreshNow
		// function will inform of the first time the currency is cached.
		rates, firstTime, err := c.refreshNow(ctx, currency)
		if err != nil {
			return nil, fmt.Errorf("refreshing cache [%v]: %w", currency, err)
		}
		if firstTime {
			c.logger.Log("msg", "scheduling periodic refresh", "currency", currency)
			go c.refreshPeriodically(ctx, currency)
		}
		return rates, nil
	}

	return rates, nil
}
