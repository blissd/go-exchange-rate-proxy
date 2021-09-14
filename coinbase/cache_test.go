package coinbase

import (
	"context"
	"github.com/stretchr/testify/assert"
	proxy "go-exchange-rate-proxy"
	"sync/atomic"
	"testing"
	"time"
)

type mock struct {
	count int32
}

func (m *mock) ExchangeRates(_ context.Context, _ proxy.Currency) (proxy.Rates, error) {
	atomic.AddInt32(&m.count, 1)
	return proxy.Rates{}, nil
}

func TestLookupWithCache(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx) // must cancel to stop go-routine started by this test
	defer cancel()

	var underlyingService mock
	s := NewCachingService(1*time.Minute, &underlyingService)

	_, _ = s.ExchangeRates(ctx, "ABC")
	assert.Equal(t, underlyingService.count, int32(1))

	_, _ = s.ExchangeRates(ctx, "ABC")
	assert.Equal(t, underlyingService.count, int32(1))
}

func TestLookupWithCache_PeriodicRefresh(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx) // must cancel to stop go-routine started by this test
	defer cancel()

	var underlyingService mock
	s := NewCachingService(1*time.Millisecond, &underlyingService)

	_, _ = s.ExchangeRates(ctx, "ABC")
	assert.GreaterOrEqual(t, underlyingService.count, int32(1))

	last := underlyingService.count
	time.Sleep(1 * time.Millisecond)
	_, _ = s.ExchangeRates(ctx, "ABC")
	assert.GreaterOrEqual(t, underlyingService.count, last)
}
