package exchange

import (
	"context"
	"go-exchange-rate-proxy"
	"reflect"
	"testing"
)

type mock struct {
	exchangeRates map[proxy.Currency]proxy.Rates
}

func (m *mock) ExchangeRates(_ context.Context, currency proxy.Currency) (proxy.Rates, error) {
	return m.exchangeRates[currency], nil
}

func TestService_Convert(t *testing.T) {
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

	cs := &mock{
		exchangeRates: allRates,
	}

	service := &service{
		coinbaseService: cs,
	}

	type args struct {
		amount proxy.Amount
		from   proxy.Currency
		to     proxy.Currency
	}
	tests := []struct {
		name    string
		args    args
		want    proxy.Exchanged
		wantErr bool
	}{
		{
			"usd -> foo",
			args{10.0, "USD", "FOO"},
			proxy.Exchanged{Rate: 2.0, Amount: 20.0},
			false,
		},
		{
			"usd -> bar",
			args{10.0, "USD", "BAR"},
			proxy.Exchanged{Rate: 3.0, Amount: 30.0},
			false,
		},
		{
			"gbp -> foo",
			args{10.0, "GBP", "FOO"},
			proxy.Exchanged{Rate: 4.0, Amount: 40.0},
			false,
		},
		{
			"gbp -> bar",
			args{10.0, "GBP", "BAR"},
			proxy.Exchanged{Rate: 5.0, Amount: 50.0},
			false,
		},
		{
			"gbp -> xyz",
			args{10.0, "GBP", "XYZ"},
			proxy.Exchanged{},
			true,
		},
		{
			"abc -> xyz",
			args{10.0, "ABC", "XYZ"},
			proxy.Exchanged{},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := service.Convert(context.Background(), tt.args.amount, tt.args.from, tt.args.to)
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
