package main

import (
	"context"
	"encoding/json"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"go-exchange-rate-proxy/coinbase"
	"go-exchange-rate-proxy/domain"
	"go-exchange-rate-proxy/proxy"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

// server dependencies for HTTP server functions
type server struct {
	proxyApi *proxy.Api
	logger   log.Logger
	router   http.ServeMux
}

func (s *server) routes() {
	s.router.Handle("/api/convert", s.convert())
}

func (s *server) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(rw, r)
}

func main() {

	w := log.NewSyncWriter(os.Stderr)
	logger := log.NewLogfmtLogger(w)
	logger = log.With(logger, "ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)
	logger = level.NewFilter(logger, level.AllowAll())
	logger = level.Debug(logger)

	cb := coinbase.New(log.With(logger, "component", "coinbase api"))
	lookup := proxy.LookupWithApi(cb)
	lookup = proxy.LookupWithCache(lookup, 1*time.Minute, logger)
	p := proxy.New(lookup, log.With(logger, "component", "proxy api"))

	server := &server{
		proxyApi: p,
		logger:   log.With(logger, "component", "http server"),
	}
	server.routes()

	http.ListenAndServe(":8080", server)
}

// convert produces HTTP handler for currency conversions
func (s *server) convert() http.HandlerFunc {

	// request for unmarshalling JSON requests posted by clients
	type request struct {
		FromCurrency domain.Currency
		ToCurrency   domain.Currency
		Amount       domain.Amount
	}

	// response for marshalling JSON responses to return to clients
	type response struct {
		Exchange domain.Rate   `json:"exchange"`
		Amount   domain.Amount `json:"amount"`
		Original domain.Amount `json:"original"`
	}

	return func(rw http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		rw.Header().Set("Content-Type", "application/json")

		bytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte(`{"error": "invalid request"}`))
			return
		}

		var request request
		err = json.Unmarshal(bytes, &request)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte(`{"error": "invalid json"}`))
			return
		}

		s.logger.Log("msg", "converting", "from", request.FromCurrency, "to", request.ToCurrency, "amount", request.Amount)

		result, err := s.proxyApi.Convert(context.Background(), request.Amount, request.FromCurrency, request.ToCurrency)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte(`{"error": "failed conversion"}`))
			return
		}

		response := response{
			Exchange: result.Rate,
			Amount:   result.Amount,
			Original: request.Amount,
		}

		enc := json.NewEncoder(rw)
		err = enc.Encode(&response)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte(`{"error": "failed json encoding"}`))
			return
		}
	}
}
