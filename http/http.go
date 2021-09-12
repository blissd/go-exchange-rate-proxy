package http

import (
	"context"
	"encoding/json"
	"github.com/go-kit/log"
	"go-exchange-rate-proxy"
	"go-exchange-rate-proxy/exchange"
	"io/ioutil"
	"net/http"
)

// Server dependencies for HTTP Server functions
type Server struct {
	Api    *exchange.Api
	Logger log.Logger
	router http.ServeMux
}

func (s *Server) Routes() {
	s.router.Handle("/api/convert", s.convert())
}

func (s *Server) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(rw, r)
}

// convert produces HTTP handler for currency conversions
func (s *Server) convert() http.HandlerFunc {

	// request for unmarshalling JSON requests posted by clients
	type request struct {
		FromCurrency proxy.Currency
		ToCurrency   proxy.Currency
		Amount       proxy.Amount
	}

	// response for marshalling JSON responses to return to clients
	type response struct {
		Exchange proxy.Rate   `json:"exchange"`
		Amount   proxy.Amount `json:"amount"`
		Original proxy.Amount `json:"original"`
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

		s.Logger.Log("msg", "converting", "from", request.FromCurrency, "to", request.ToCurrency, "amount", request.Amount)

		result, err := s.Api.Convert(context.Background(), request.Amount, request.FromCurrency, request.ToCurrency)
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