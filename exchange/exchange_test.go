package exchange

import (
	"context"
	"errors"
	"github.com/go-kit/log"
	"github.com/stretchr/testify/assert"
	"go-exchange-rate-proxy"
	"reflect"
	"sync/atomic"
	"testing"
	"time"
)

func TestLookupWithCache(t *testing.T) {
	invocationCount := 0

	var underlyingLookup LookupFunc = func(ctx context.Context, currency proxy.Currency) (proxy.Rates, error) {
		invocationCount++
		return proxy.Rates{}, nil
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx) // must cancel to stop go-routine started by this test
	defer cancel()

	cachedLookup := LookupWithCache(underlyingLookup, 1*time.Minute, log.NewNopLogger())

	_, _ = cachedLookup(ctx, "ABC")
	assert.Equal(t, invocationCount, 1)

	_, _ = cachedLookup(ctx, "ABC")
	assert.Equal(t, invocationCount, 1)
}

func TestLookupWithCache_PeriodicRefresh(t *testing.T) {
	var invocationCount int32

	var underlyingLookup LookupFunc = func(ctx context.Context, currency proxy.Currency) (proxy.Rates, error) {
		invocationCount++
		atomic.AddInt32(&invocationCount, 1)
		return proxy.Rates{}, nil
	}

	cachedLookup := LookupWithCache(underlyingLookup, 1*time.Millisecond, log.NewNopLogger())

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx) // must cancel to stop go-routine started by this test
	defer cancel()

	_, _ = cachedLookup(ctx, "ABC")
	assert.GreaterOrEqual(t, invocationCount, int32(1))

	last := invocationCount
	time.Sleep(1 * time.Millisecond)
	_, _ = cachedLookup(ctx, "ABC")
	assert.GreaterOrEqual(t, invocationCount, last)
}

func TestApi_Convert(t *testing.T) {

	usdRates := proxy.Rates{
		"FOO": 2.0,
		"BAR": 3.0,
	}

	gbpRates := proxy.Rates{
		"FOO": 4.0,
		"BAR": 5.0,
	}

	allRates := map[proxy.Currency]proxy.Rates{
		"USD": usdRates,
		"GBP": gbpRates,
	}

	var lookup LookupFunc = func(ctx context.Context, currency proxy.Currency) (proxy.Rates, error) {
		rates, ok := allRates[currency]
		if !ok {
			return nil, errors.New("bad rate")
		}
		return rates, nil
	}

	api := Api{
		lookup: lookup,
		logger: log.NewNopLogger(),
	}

	type args struct {
		amount proxy.Amount
		from   proxy.Currency
		to     proxy.Currency
	}
	tests := []struct {
		name    string
		args    args
		want    *proxy.Exchanged
		wantErr bool
	}{
		{
			"usd -> foo",
			args{10.0, "USD", "FOO"},
			&proxy.Exchanged{Rate: 2.0, Amount: 20.0},
			false,
		},
		{
			"usd -> bar",
			args{10.0, "USD", "BAR"},
			&proxy.Exchanged{Rate: 3.0, Amount: 30.0},
			false,
		},
		{
			"gbp -> foo",
			args{10.0, "GBP", "FOO"},
			&proxy.Exchanged{Rate: 4.0, Amount: 40.0},
			false,
		},
		{
			"gbp -> bar",
			args{10.0, "GBP", "BAR"},
			&proxy.Exchanged{Rate: 5.0, Amount: 50.0},
			false,
		},
		{
			"gbp -> xyz",
			args{10.0, "GBP", "XYZ"},
			nil,
			true,
		},
		{
			"abc -> xyz",
			args{10.0, "ABC", "XYZ"},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := api.Convert(context.Background(), tt.args.amount, tt.args.from, tt.args.to)
			if (err != nil) != tt.wantErr {
				t.Errorf("Convert() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Convert() got = %v, want %v", got, tt.want)
			}
		})
	}
}