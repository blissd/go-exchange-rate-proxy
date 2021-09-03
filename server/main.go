package main

import (
	"encoding/json"
	"github.com/go-kit/log"
	"go-exchange-rate-proxy/coinbase"
	"go-exchange-rate-proxy/domain"
	"go-exchange-rate-proxy/proxy"
	"io/ioutil"
	"net/http"
	"os"
)

type server struct {
	proxyApi *proxy.Api
	logger   log.Logger
}

func main() {
	w := log.NewSyncWriter(os.Stderr)
	logger := log.NewLogfmtLogger(w)

	cb := coinbase.New(log.With(logger, "component", "coinbase api"))
	p := proxy.New(cb, log.With(logger, "component", "proxy api"))

	server := &server{
		proxyApi: p,
		logger:   log.With(logger, "component", "http server"),
	}

	http.Handle("/api/convert", http.Handler(server.convert()))
	http.ListenAndServe(":8080", nil)
}

type ConvertRequest struct {
	FromCurrency domain.Currency
	ToCurrency   domain.Currency
	Amount       domain.Amount
}

type ConvertResponse struct {
	Exchange domain.Rate
	Amount   domain.Amount
	Original domain.Amount
}

func (s *server) convert() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		rw.Header().Set("Content-Type", "application/json")

		bytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte(`{"error": "invalid request"}`))
			return
		}

		var request ConvertRequest
		err = json.Unmarshal(bytes, &request)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte(`{"error": "invalid json"}`))
			return
		}

		s.logger.Log("msg", "converting", "from", request.FromCurrency, "to", request.ToCurrency, "amount", request.Amount)

		result, err := s.proxyApi.Convert(request.Amount, request.FromCurrency, request.ToCurrency)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte(`{"error": "failed conversion"}`))
			return
		}

		response := ConvertResponse{
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
