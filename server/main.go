package main

import (
	"github.com/go-kit/log"
	"go-exchange-rate-proxy/coinbase"
	"go-exchange-rate-proxy/convert"
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

	coinbaseService := coinbase.NewService()
	coinbaseService = coinbase.NewLoggingService(log.With(logger, "component", "coinbase_rest"), coinbaseService)
	coinbaseService = coinbase.NewCachingService(1*time.Minute, log.With(logger, "component", "coinbase_cache"), coinbaseService)
	coinbaseService = coinbase.NewLoggingService(log.With(logger, "component", "coinbase_cache"), coinbaseService)

	convertService := convert.NewService(coinbaseService)
	convertService = convert.NewLoggingService(log.With(logger, "component", "convert"), convertService)

	handler := http.NewHandler(convertService)
	nhttp.ListenAndServe(":8080", handler)
}
