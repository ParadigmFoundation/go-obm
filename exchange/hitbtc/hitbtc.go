package hitbtc

import (
	"context"
	"log"
	"strings"
	"sync"

	"github.com/ParadigmFoundation/go-obm"
	"github.com/ParadigmFoundation/go-obm/exchange"
	"github.com/bitbandi/go-hitbtc"
)

type Exchange struct {
}

func New() *Exchange { return &Exchange{} }

func (x *Exchange) Subscribe(ctx context.Context, sub exchange.Subscriber, syms ...string) error {
	client, err := hitbtc.NewWSClient()
	if err != nil {
		return err
	}
	defer client.Close()

	wg := sync.WaitGroup{}
	for _, sym := range syms {
		sym := sym
		fmtSym := strings.Replace(sym, "/", "", 1)
		log.Printf("Adding: %s", fmtSym)
		updates, snapshot, err := client.SubscribeOrderbook(fmtSym)
		if err != nil {
			return err
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			subscribe(sub, sym, snapshot, updates)
		}()
	}
	wg.Wait()

	return nil
}

func subscribe(sub exchange.Subscriber, sym string,
	snapshot <-chan hitbtc.WSNotificationOrderbookSnapshot,
	updates <-chan hitbtc.WSNotificationOrderbookUpdate) {
	name := "hitbtc"
	for {
		select {
		case snapshot := <-snapshot:
			up, err := newUpdate(sym, snapshot.Ask, snapshot.Bid)
			if err != nil {
				log.Printf("err: %+v", err)
			}
			if err := sub.OnSnapshot(name, up); err != nil {
				log.Printf("err: %+v", err)
			}
		case update := <-updates:
			up, err := newUpdate(sym, update.Ask, update.Bid)
			if err != nil {
				log.Printf("err: %+v", err)
				continue
			}

			if err := sub.OnUpdate(name, up); err != nil {
				log.Printf("err: %+v", err)
			}
		}
	}
}

func newUpdate(symbol string, asks, bids []hitbtc.WSSubtypeTrade) (*obm.Update, error) {
	update := &obm.Update{
		Symbol: symbol,
	}

	for _, ask := range asks {
		entry, err := obm.NewEntryFromStrings(ask.Price, ask.Size)
		if err != nil {
			return nil, err
		}
		update.Asks = append(update.Asks, entry)
	}

	for _, bid := range bids {
		entry, err := obm.NewEntryFromStrings(bid.Price, bid.Size)
		if err != nil {
			return nil, err
		}
		update.Bids = append(update.Bids, entry)
	}

	return update, nil
}
