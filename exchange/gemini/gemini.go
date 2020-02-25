package gemini

import (
	"context"
	"log"
	"strings"
	"sync"

	"github.com/ParadigmFoundation/go-obm"
	"github.com/ParadigmFoundation/go-obm/exchange"
	"github.com/gorilla/websocket"
)

const (
	NAME     = "gemini"
	FEED_URL = "wss://api.sandbox.gemini.com/v1/marketdata/"
)

type Exchange struct {
	sMutex  sync.RWMutex
	symbols map[string]string
}

func New() *Exchange {
	return &Exchange{
		symbols: make(map[string]string),
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
	for i := range syms {
		syms[i] = x.newSymbol(syms[i])
	}
	log.Printf("Gemini querying: %q", syms)

	errCh := make(chan error)
	for _, sym := range syms {
		sym := sym
		go func() {
			errCh <- x.subscribe(ctx, sub, sym)
		}()
	}

	return <-errCh
}

func (x *Exchange) subscribe(ctx context.Context, sub exchange.Subscriber, sym string) error {
	url := FEED_URL + sym //+ "?trades=false&auctions=false"
	c, _, err := websocket.DefaultDialer.DialContext(ctx, url, nil)
	if err != nil {
		return nil
	}
	defer c.Close()

	errCh := make(chan error, 1)
	go func() {
		defer close(errCh)

		for {
			errCh <- x.handleWs(c, sub, sym)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-errCh:
			if err != nil {
				return err
			}
		}
	}
}

func (x *Exchange) handleWs(c *websocket.Conn, sub exchange.Subscriber, sym string) error {
	var msg Message
	err := c.ReadJSON(&msg)
	if err != nil {
		return err
	}

	if msg.Type != "update" {
		return nil
	}

	update := &obm.Update{
		Symbol: x.symbol(sym),
	}

	for _, event := range msg.Events {
		if event.Type != "change" {
			continue
		}

		entry, err := obm.NewEntryFromStrings(event.Price, event.Remaining)
		if err != nil {
			return err
		}

		switch event.Side {
		case "bid":
			update.Bids = append(update.Bids, entry)
		case "ask":
			update.Asks = append(update.Asks, entry)
		}
	}

	if len(update.Bids) != 0 || len(update.Asks) != 0 {
		return sub.OnUpdate(NAME, update)
	}

	return nil
}

func init() { exchange.Register(NAME, New()) }
