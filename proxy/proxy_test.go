package proxy

import (
	"github.com/go-kit/log"
	"github.com/stretchr/testify/assert"
	"go-exchange-rate-proxy/domain"
	"sync/atomic"
	"testing"
	"time"
)

func TestLookupWithCache(t *testing.T) {
	invocationCount := 0

	var underlyingLookup LookupFunc = func(currency domain.Currency) (domain.Rates, error) {
		invocationCount++
		return domain.Rates{}, nil
	}

	cachedLookup := LookupWithCache(underlyingLookup, 1*time.Minute, log.NewNopLogger())

	_, _ = cachedLookup("ABC")
	assert.Equal(t, invocationCount, 1)

	_, _ = cachedLookup("ABC")
	assert.Equal(t, invocationCount, 1)
}

func TestLookupWithCache_PeriodicRefresh(t *testing.T) {
	var invocationCount int32

	var underlyingLookup LookupFunc = func(currency domain.Currency) (domain.Rates, error) {
		invocationCount++
		atomic.AddInt32(&invocationCount, 1)
		return domain.Rates{}, nil
	}

	cachedLookup := LookupWithCache(underlyingLookup, 1*time.Millisecond, log.NewNopLogger())

	_, _ = cachedLookup("ABC")
	assert.GreaterOrEqual(t, invocationCount, int32(1))

	last := invocationCount
	time.Sleep(1 * time.Millisecond)
	_, _ = cachedLookup("ABC")
	assert.GreaterOrEqual(t, invocationCount, last)
}
