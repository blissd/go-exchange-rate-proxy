package exchange

import (
	"context"
	"github.com/go-kit/log"
	"go-exchange-rate-proxy"
	"time"
)

// loggingService decorates an exchange.Service with logging
type loggingService struct {
	logger log.Logger
	next   Service
}

// NewLoggingService returns a new instance of a logging Service
func NewLoggingService(logger log.Logger, s Service) Service {
	return &loggingService{
		next:   s,
		logger: logger,
	}
}

func (s *loggingService) Convert(ctx context.Context, amount proxy.Amount, from proxy.Currency, to proxy.Currency) (ex proxy.Exchanged, err error) {
	defer func(begin time.Time) {
		s.logger.Log(
			"method", "convert",
			"amount", amount,
			"from", from,
			"to", to,
			"rate", ex.Rate,
			"converted_amount", ex.Amount,
			"took", time.Since(begin),
			"err", err,
		)
	}(time.Now())
	return s.next.Convert(ctx, amount, from, to)
}
