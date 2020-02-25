package exchange

import (
	"context"
	"errors"

	"github.com/ParadigmFoundation/go-obm"
)

type Subscriber interface {
	OnSnapshot(string, *obm.Update) error
	OnUpdate(string, *obm.Update) error
}

type Exchange interface {
	// Subscribe subscribes to one or more symbols with a given set of callbacks
	Subscribe(context.Context, Subscriber, ...string) error
}

var exchanges = make(map[string]Exchange)

func Register(name string, x Exchange) { exchanges[name] = x }

var ErrExchangeNotRegistered = errors.New("exchange not registered")

func Get(name string) (Exchange, error) {
	x, ok := exchanges[name]
	if !ok {
		return nil, ErrExchangeNotRegistered
	}
	return x, nil
}

func MustGet(name string) Exchange {
	x, err := Get(name)
	if err != nil {
		panic(err)
	}
	return x
}

func ListNames() []string {
	var names []string
	for name := range exchanges {
		names = append(names, name)
	}
	return names
}
