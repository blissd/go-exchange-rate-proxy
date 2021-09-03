package main

import (
	"github.com/go-kit/log"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCoinbaseApi_ExchangeRates(t *testing.T) {
	response := `{
		"data": {
			"currency": "USD",
			"rates": {
				"BCH": "1000.0",
				"GBP": "1.2"
			}
		}
	}`

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.True(t, strings.HasSuffix(req.URL.String(), "exchange-rates?currency=USD"))
		// Send response to be tested
		rw.Write([]byte(response))
	}))
	defer server.Close()

	api := CoinbaseApi{
		url:    server.URL,
		logger: log.NewNopLogger(),
	}

	rates, err := api.ExchangeRates("USD")

	assert.Nil(t, err)
	assert.Equal(t, Rate(1000.0), rates["BCH"])
	assert.Equal(t, Rate(1.2), rates["GBP"])
}
