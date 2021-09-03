package domain

// Currency a currency code
type Currency string

// Amount a monetary amount... which should be a float...
type Amount float64

type Exchanged struct {
	Rate   Rate
	Amount Amount
}

// Rate an exchange rate
type Rate float64

type Rates map[Currency]Rate
