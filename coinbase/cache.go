package coinbase

import (
	"context"
	"fmt"
	"github.com/go-kit/log"
	"go-exchange-rate-proxy"
	"sync"
	"time"
)

// cachingService decorates a coinbase.Service with a cache of exchange rates.
// The cachingService is concurrency safe and will periodically refresh cached values.
type cachingService struct {
	// next the service being decorated with a cache
	next Service
	// cache the cache of rates
	cache map[proxy.Currency]proxy.Rates

	// updateFrequency how often to refresh cached values
	updateFrequency time.Duration

	// lock synchronizes access to cache to make it concurrency safe
	lock sync.RWMutex

	logger log.Logger
}

// NewCachingService returns a new caching Service
func NewCachingService(updateFrequency time.Duration, s Service) Service {
	return &cachingService{
		next:            s,
		cache:           map[proxy.Currency]proxy.Rates{},
		updateFrequency: updateFrequency,
		lock:            sync.RWMutex{},
		logger:          log.NewNopLogger(),
	}
}

// ExchangeRates looks up exchange rates and caches the results
func (s *cachingService) ExchangeRates(ctx context.Context, currency proxy.Currency) (proxy.Rates, error) {
	s.lock.RLock()
	rates, ok := s.cache[currency]
	s.lock.RUnlock()

	if !ok {
		// Note there is a race condition here in that multiple requests for a currency that isn't yet cached
		// will result in multiple concurrent attempts to refresh. This should be harmless, unless the underlying
		// coinbase API throttles the requests. We could avoid this by holding a lock while calling the
		// calling and waiting on the underlying coinbase API, but that is a blocking operation so I'd rather not.
		// To avoid running multiple go routines to periodically refresh the same currency, the refreshNow
		// function will inform of the first time the currency is cached.
		rates, firstTime, err := s.refreshNow(ctx, currency)
		if err != nil {
			return nil, fmt.Errorf("refreshing cachingService [%v]: %w", currency, err)
		}
		if firstTime {
			go s.refreshPeriodically(ctx, currency)
		}
		return rates, nil
	}

	return rates, nil
}

// refreshNow refreshes a cached entry immediately
func (s *cachingService) refreshNow(ctx context.Context, currency proxy.Currency) (proxy.Rates, bool, error) {
	rates, err := s.next.ExchangeRates(ctx, currency)
	if err != nil {
		return nil, false, fmt.Errorf("refresh [%v]: %w", currency, err)
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	_, ok := s.cache[currency]
	s.cache[currency] = rates
	return rates, !ok, nil
}

// refreshPeriodically refreshes a cached entry on a given schedule.
// This is expected to be called from a go-routine for each currency.
func (s *cachingService) refreshPeriodically(ctx context.Context, currency proxy.Currency) {
	for {
		select {
		case <-time.After(s.updateFrequency):
			_, _, err := s.refreshNow(ctx, currency)
			if err != nil {
				// Don't return, just log and hope this is a transient error
				s.logger.Log("msg", "periodic refresh failed", "currency", currency, "error", err)
			}
		case <-ctx.Done():
			s.uncache(currency)
			return
		}
	}
}

// uncache safely removes currency from cachingService
func (s *cachingService) uncache(currency proxy.Currency) {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.cache, currency)
}
