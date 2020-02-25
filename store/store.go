package store

import (
	"github.com/ParadigmFoundation/go-obm/exchange"
	"github.com/ParadigmFoundation/go-obm/grpc/types"
)

type Store interface {
	exchange.Subscriber
	OrderBook(exchange, symbol string) (*types.OrderBookResponse, error)
}
