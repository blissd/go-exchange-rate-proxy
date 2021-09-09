package main

import (
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"go-exchange-rate-proxy/coinbase"
	"go-exchange-rate-proxy/exchange"
	"go-exchange-rate-proxy/http"
	"os"
	"time"

	nhttp "net/http"
)

func main() {
	w := log.NewSyncWriter(os.Stderr)
	logger := log.NewLogfmtLogger(w)
	logger = log.With(logger, "ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)
	logger = level.NewFilter(logger, level.AllowAll())
	logger = level.Debug(logger)

	cb := coinbase.New(log.With(logger, "component", "coinbase api"))
	lookup := exchange.LookupWithApi(cb)
	lookup = exchange.LookupWithCache(lookup, 1*time.Minute, logger)
	p := exchange.New(lookup, log.With(logger, "component", "proxy api"))

	server := &http.Server{
		Proxy:  p,
		Logger: log.With(logger, "component", "http server"),
	}
	server.Routes()

	nhttp.ListenAndServe(":8080", server)
}
