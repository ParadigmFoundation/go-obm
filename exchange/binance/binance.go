package binance

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/ParadigmFoundation/go-obm"
	"github.com/ParadigmFoundation/go-obm/exchange"
	"github.com/adshao/go-binance"
)

const (
	NAME = "binance"
)

/*
How to manage a local order book correctly:
https://github.com/binance-exchange/binance-official-api-docs/blob/master/web-socket-streams.md#how-to-manage-a-local-order-book-correctly

1. Open a stream to wss://stream.binance.com:9443/ws/bnbbtc@depth.
2. Buffer the events you receive from the stream.
3. Get a depth snapshot from binance.com/api/v1/depth?symbol=BNBBTC&limit=1000 .
4. Drop any event where u is <= lastUpdateId in the snapshot.
5. The first processed event should have U <= lastUpdateId+1 AND u >= lastUpdateId+1.
6. While listening to the stream, each new event's U should be equal to the previous event's u+1.
7. The data in each event is the absolute quantity for a price level.
8. If the quantity is 0, remove the price level.
9. Receiving an event that removes a price level that is not in your local order book can happen and is normal.
*/
type Exchange struct {
	sMutex  sync.RWMutex
	symbols map[string]string

	// events
	events chan *binance.WsDepthEvent
}

func New() *Exchange {
	return &Exchange{
		symbols: make(map[string]string),
		events:  make(chan *binance.WsDepthEvent, 1000),
	}
}

func (x *Exchange) handleEvent(symbol string, sub exchange.Subscriber) error {
	// we use depth to get the snapshot, id  and secret
	depth := binance.NewClient("", "").NewDepthService()
	depth.Symbol(symbol)
	depth.Limit(1000)

	res, err := depth.Do(context.Background())
	if err != nil {
		return err
	}

	// Turn the response into events so that we can reuse our newUpdates function
	updates, err := x.newUpdates(&binance.WsDepthEvent{
		Symbol: symbol,
		Asks:   res.Asks,
		Bids:   res.Bids,
	})
	if err != nil {
		return err
	}
	if err := sub.OnSnapshot(NAME, updates); err != nil {
		return err
	}

	// read events and update the Subscriber
	for {
		event := <-x.events
		if event.UpdateID <= res.LastUpdateID {
			log.Printf("skipping event %d / %d", event.UpdateID, res.LastUpdateID)
			continue
		}
		update, err := x.newUpdates(event)
		if err != nil {
			return err
		}
		if err := sub.OnUpdate(NAME, update); err != nil {
			return err
		}
	}
}

func (x *Exchange) newSymbol(s string) string {
	x.sMutex.Lock()
	defer x.sMutex.Unlock()

	newSymbol := strings.Replace(s, "/", "", 1)
	x.symbols[newSymbol] = s
	return newSymbol
}

func (x *Exchange) symbol(s string) string {
	x.sMutex.RLock()
	defer x.sMutex.RUnlock()

	if found := x.symbols[s]; found != "" {
		return found
	}
	return s
}

func (x *Exchange) Subscribe(ctx context.Context, sub exchange.Subscriber, syms ...string) error {
	eventHandler := func(event *binance.WsDepthEvent) {
		x.events <- event
	}

	errHandler := func(err error) {
		log.Printf("ERROR: %+v", err)
	}

	for i := range syms {
		syms[i] = x.newSymbol(syms[i])
	}
	log.Printf("Binance querying: %q", syms)
	for _, sym := range syms {
		_, stop, err := binance.WsDepthServe(sym, eventHandler, errHandler)
		if err != nil {
			return fmt.Errorf("Subscribe(): %w", err)
		}

		errCh := make(chan error, 1)
		go func() {
			for {
				errCh <- x.handleEvent(sym, sub)
			}
		}()

		go func() {
			select {
			case <-ctx.Done():
				stop <- struct{}{}
			case err := <-errCh:
				log.Printf("err = %+v\n", err)
			}
		}()
	}

	<-ctx.Done()
	return ctx.Err()
}

func (x *Exchange) newUpdates(event *binance.WsDepthEvent) (*obm.Update, error) {
	var updates = obm.Update{
		Symbol: x.symbol(event.Symbol),
	}

	for _, bid := range event.Bids {
		entry, err := obm.NewEntryFromStrings(bid.Price, bid.Quantity)
		if err != nil {
			return nil, err
		}
		updates.Bids = append(updates.Bids, entry)
	}

	for _, ask := range event.Asks {
		entry, err := obm.NewEntryFromStrings(ask.Price, ask.Quantity)
		if err != nil {
			return nil, err
		}
		updates.Asks = append(updates.Asks, entry)
	}

	return &updates, nil
}

func init() {
	exchange.Register(NAME, New())
}
