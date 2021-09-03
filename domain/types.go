package domain

// Currency a currency code
type Currency string

// Amount a monetary amount... which should not be a float...
type Amount float64

// Exchanged result of computing a currency conversion
type Exchanged struct {
	// Rate exchange rate used for conversion
	Rate Rate

	// Amount converted amount
	Amount Amount
}

// Rate an exchange rate
type Rate float64

// Rate exchange rates for currencies
type Rates map[Currency]Rate
