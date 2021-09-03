package main

import (
	"github.com/go-kit/log"
	"go-exchange-rate-proxy/coinbase"
	"go-exchange-rate-proxy/proxy"
	"os"
)

func main() {
	w := log.NewSyncWriter(os.Stderr)
	logger := log.NewLogfmtLogger(w)

	cb := coinbase.New(logger)
	p := proxy.New(cb, logger)
	convertedAmount, err := p.Convert(1.0, "GBP", "BCH")
	if err != nil {
		logger.Log("msg", "failed conversion", "error", err)
		os.Exit(1)
	}

	logger.Log("msg", "converted", "amount", convertedAmount)

}
