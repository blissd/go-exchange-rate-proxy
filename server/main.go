package main

import (
	"github.com/go-kit/log"
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
	//logger = level.NewFilter(logger, level.AllowAll())
	//logger = level.Debug(logger)

	cs := coinbase.NewService()
	cs = coinbase.NewLoggingService(log.With(logger, "component", "coinbase_rest"), cs)
	cs = coinbase.NewCachingService(1*time.Minute, cs)
	cs = coinbase.NewLoggingService(log.With(logger, "component", "coinbase_cache"), cs)

	es := exchange.NewService(cs)
	es = exchange.NewLoggingService(log.With(logger, "component", "exchange"), es)

	handler := http.NewHandler(es)
	nhttp.ListenAndServe(":8080", handler)
}
