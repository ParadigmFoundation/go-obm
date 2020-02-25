package obm

import (
	"strconv"
)

type Entry struct {
	Price    float64 `json:"price"`
	Quantity float64 `json:"quantity"`
}

func NewEntryFromStrings(p, q string) (*Entry, error) {
	price, err := strconv.ParseFloat(p, 64)
	if err != nil {
		return nil, err
	}

	quantity, err := strconv.ParseFloat(q, 64)
	if err != nil {
		return nil, err
	}

	return &Entry{Price: price, Quantity: quantity}, nil
}

type Entries []*Entry

type Update struct {
	Symbol string
	Bids   Entries
	Asks   Entries
}
