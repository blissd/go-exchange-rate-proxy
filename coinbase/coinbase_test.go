package coinbase

import (
	"context"
	"github.com/go-kit/log"
	"github.com/stretchr/testify/assert"
	"go-exchange-rate-proxy"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestCoinbaseApi_ExchangeRates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.True(t, strings.HasSuffix(req.URL.String(), "/exchange-rates?currency=USD"))
		response := `{
			"data": {
				"currency": "USD",
				"rates": {
					"BCH": "1000.0",
					"GBP": "1.2"
				}
			}
		}`
		_, _ = rw.Write([]byte(response))
	}))
	defer server.Close()

	api := Api{
		url:    server.URL,
		logger: log.NewNopLogger(),
	}

	rates, err := api.ExchangeRates(context.Background(), "USD")

	assert.Nil(t, err)
	assert.Equal(t, proxy.Rate(1000.0), rates["BCH"])
	assert.Equal(t, proxy.Rate(1.2), rates["GBP"])
}

func TestCoinbaseApi_ExchangeRatesTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		time.Sleep(10 * time.Millisecond)
		_, _ = rw.Write([]byte("{}"))
	}))
	defer server.Close()

	api := Api{
		url:    server.URL,
		logger: log.NewNopLogger(),
	}
	api.client.Timeout = 1 * time.Millisecond

	_, err := api.ExchangeRates(context.Background(), "USD")

	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "Client.Timeout")) // fragile :-(
}
