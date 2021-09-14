package exchange

import (
	"context"
	"fmt"
	"go-exchange-rate-proxy"
	"go-exchange-rate-proxy/coinbase"
)

// Service interface for converting from one interface to another
type Service interface {
	Convert(ctx context.Context, amount proxy.Amount, from proxy.Currency, to proxy.Currency) (proxy.Exchanged, error)
}

// Service proxy exchange API
type service struct {
	// coinbase coinbaseService to lookup exchange exchangeRates.
	coinbaseService coinbase.Service
}

// NewService constructs a valid Service
func NewService(s coinbase.Service) Service {
	return &service{
		coinbaseService: s,
	}
}

// Convert computes a conversion from one currency to another with the current exchange rate.
// As a side-effect the cache of exchange exchangeRates might be updated.
func (s *service) Convert(ctx context.Context, amount proxy.Amount, from proxy.Currency, to proxy.Currency) (proxy.Exchanged, error) {
	rates, err := s.coinbaseService.ExchangeRates(ctx, from)
	if err != nil {
		return proxy.Exchanged{}, fmt.Errorf("convert from [%v]: %w", from, err)
	}

	rate, ok := rates[to]
	if !ok {
		return proxy.Exchanged{}, fmt.Errorf("unknown 'to' currency: %v", to)
	}

	result := proxy.Exchanged{
		Rate:   rate,
		Amount: proxy.Amount(float64(rate) * float64(amount)),
	}

	return result, nil
}
