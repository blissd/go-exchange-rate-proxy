package http

import (
	"context"
	"github.com/stretchr/testify/assert"
	proxy "go-exchange-rate-proxy"
	"net/http/httptest"
	"strings"
	"testing"
)

type mock struct {
	t      *testing.T
	amount proxy.Amount
	from   proxy.Currency
	to     proxy.Currency
}

func (m *mock) Convert(_ context.Context, amount proxy.Amount, from proxy.Currency, to proxy.Currency) (proxy.Exchanged, error) {
	assert.Equal(m.t, m.amount, amount, "amount")
	assert.Equal(m.t, m.from, from, "from")
	assert.Equal(m.t, m.to, to, "to")
	return proxy.Exchanged{Rate: 2.0, Amount: 6.0}, nil
}

func TestServer_ServeHTTP(t *testing.T) {
	es := mock{
		t:      t,
		amount: 3,
		from:   "GBP",
		to:     "FOO",
	}

	server := NewServer(&es)

	w := httptest.NewRecorder()
	msg := `{"fromCurrency":"GBP", "toCurrency":"FOO","amount":3.0}`
	r := httptest.NewRequest("POST", "/api/convert", strings.NewReader(msg))

	server.ServeHTTP(w, r)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, `{"exchange":2,"amount":6,"original":3}`, strings.TrimSpace(w.Body.String()))
}
