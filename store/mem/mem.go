package mem

import (
	"sort"
	"sync"
	"time"

	"github.com/ParadigmFoundation/go-obm"
	"github.com/ParadigmFoundation/go-obm/grpc/types"
)

type market struct {
	lastUpdate time.Time
	bids       *CappedTree
	asks       *CappedTree
}

// these operates as db indexes
type markets map[string]*market
type exchanges map[string]markets

type Store struct {
	xch exchanges
	m   sync.RWMutex
}

func New() *Store {
	return &Store{
		xch: make(map[string]markets),
	}
}

func (s *Store) OnSnapshot(name string, update *obm.Update) error {
	return s.doUpdate(name, update)
}

func (s *Store) OnUpdate(name string, update *obm.Update) error {
	return s.doUpdate(name, update)
}

func (s *Store) findOrCreateMarket(name, symbol string) *market {
	if s.xch[name] == nil {
		s.xch[name] = make(map[string]*market)
	}

	mkt := s.xch[name][symbol]
	if mkt == nil {
		mkt = &market{
			bids: NewCappedTree(200, RemoveMin),
			asks: NewCappedTree(200, RemoveMax),
		}
		s.xch[name][symbol] = mkt
	}

	return mkt
}

func (s *Store) doUpdate(name string, update *obm.Update) error {
	s.m.Lock()
	defer s.m.Unlock()

	mkt := s.findOrCreateMarket(name, update.Symbol)

	for _, bid := range update.Bids {
		p, q := bid.Price, bid.Quantity
		if q == 0 {
			mkt.bids.Remove(p)
		} else {
			mkt.bids.Put(p, q)
		}
	}

	for _, ask := range update.Asks {
		p, q := ask.Price, ask.Quantity
		if q == 0 {
			mkt.asks.Remove(p)
		} else {
			mkt.asks.Put(p, q)
		}
	}

	mkt.lastUpdate = time.Now()
	return nil
}

func (s *Store) OrderBook(exchange, symbol string) (*types.OrderBookResponse, error) {
	s.m.RLock()
	defer s.m.RUnlock()

	mkt := s.findOrCreateMarket(exchange, symbol)

	var asks types.OrderBookEntriesByPriceAsc
	mkt.asks.Each(func(key, val float64) {
		asks = append(asks, &types.OrderBookEntry{Price: key, Quantity: val})
	})
	sort.Sort(asks)

	var bids types.OrderBookEntriesByPriceDesc
	mkt.bids.Each(func(key, val float64) {
		bids = append(bids, &types.OrderBookEntry{Price: key, Quantity: val})
	})
	sort.Sort(bids)

	ob := &types.OrderBookResponse{
		Exchange: exchange,
		Symbol:   symbol,
		Asks:     asks,
		Bids:     bids,
	}
	if !mkt.lastUpdate.IsZero() {
		ob.LastUpdate = mkt.lastUpdate.Unix()
	}

	return ob, nil
}
