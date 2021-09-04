package proxy

import (
	"github.com/go-kit/log"
	"github.com/stretchr/testify/assert"
	"go-exchange-rate-proxy/domain"
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
