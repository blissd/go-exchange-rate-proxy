package main

import (
	"github.com/go-kit/log"
	"os"
)

// Currency a currency code
type Currency string

// Amount a monetary amount... which should be a float...
type Amount float64

// Rate an exchange rate
type Rate float64

type Rates map[Currency]Rate

func main() {

	w := log.NewSyncWriter(os.Stderr)
	logger := log.NewLogfmtLogger(w)

	coinbase := NewCoinbaseApi(logger)

	var currency Currency = "USD"
	rates, err := coinbase.ExchangeRates(currency)
	if err != nil {
		logger.Log("msg", err)
		panic(-1)
	}

	logger.Log("msg", "exchange rates", "currency", currency, "rate", rates["BCH"])

}
