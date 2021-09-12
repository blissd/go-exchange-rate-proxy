package http

import (
	"context"
	"errors"
	"github.com/go-kit/log"
	"github.com/stretchr/testify/assert"
	proxy "go-exchange-rate-proxy"
	"go-exchange-rate-proxy/exchange"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestServer_ServeHTTP(t *testing.T) {
	gbpRates := proxy.Rates{
		"FOO": 4.0,
		"BAR": 5.0,
	}

	allRates := map[proxy.Currency]proxy.Rates{
		"GBP": gbpRates,
	}

	var lookup exchange.LookupFunc = func(ctx context.Context, currency proxy.Currency) (proxy.Rates, error) {
		rates, ok := allRates[currency]
		if !ok {
			return nil, errors.New("bad rate")
		}
		return rates, nil
	}

	api := exchange.New(lookup, log.NewNopLogger())

	server := Server{
		Api:    api,
		Logger: log.NewNopLogger(),
	}
	server.Routes()

	w := httptest.NewRecorder()
	msg := `{"fromCurrency":"GBP", "toCurrency":"FOO","amount":2.0}`
	r := httptest.NewRequest("POST", "/api/convert", strings.NewReader(msg))

	server.ServeHTTP(w, r)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, `{"exchange":4,"amount":8,"original":2}`, strings.TrimSpace(w.Body.String()))
}
