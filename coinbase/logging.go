package coinbase

import (
	"context"
	"github.com/go-kit/log"
	"go-exchange-rate-proxy"
	"time"
)

// loggingService decorates a coinbase.Service with logging
type loggingService struct {
	next   Service
	logger log.Logger
}

// NewLoggingService return a new logging service
func NewLoggingService(logger log.Logger, s Service) Service {
	return &loggingService{
		next:   s,
		logger: logger,
	}
}

func (s *loggingService) ExchangeRates(ctx context.Context, currency proxy.Currency) (rates proxy.Rates, err error) {
	defer func(begin time.Time) {
		s.logger.Log(
			"method", "exchange_rates",
			"currency", currency,
			"took", time.Since(begin),
			"err", err,
		)
	}(time.Now())
	return s.next.ExchangeRates(ctx, currency)
}
