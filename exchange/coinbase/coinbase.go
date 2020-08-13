package coinbase

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	coinbasepro "github.com/preichenberger/go-coinbasepro/v2"

	"github.com/ParadigmFoundation/go-obm"
	"github.com/ParadigmFoundation/go-obm/exchange"
)

const (
	NAME     = "coinbase"
	FEED_URL = "wss://ws-feed.pro.coinbase.com"
)

var _ exchange.Exchange = &Exchange{}

type Exchange struct {
	url        string
	retryAfter time.Duration
}

type Option func(x *Exchange)

func WithURL(url string) Option {
	return func(x *Exchange) { x.url = url }
}

func WithRetryAfter(d time.Duration) Option {
	return func(x *Exchange) { x.retryAfter = d }
}

func New(options ...Option) *Exchange {
	x := &Exchange{
		url:        FEED_URL,
		retryAfter: 1 * time.Second,
	}

	for _, opt := range options {
		opt(x)
	}
	return x
}

func (x *Exchange) dial(ctx context.Context) (*websocket.Conn, error) {
	var ws websocket.Dialer
	conn, _, err := ws.DialContext(ctx, x.url, nil)
	if err != nil {
		return nil, fmt.Errorf("Coinbase: DialContext(): %w", err)
	}
	return conn, nil
}

func symbol2coinbase(s string) string { return strings.Replace(s, "/", "-", 1) }
func coinbase2symbol(s string) string { return strings.Replace(s, "-", "/", 1) }

func (x *Exchange) Subscribe(ctx context.Context, sub exchange.Subscriber, syms ...string) error {
	for {
		err := x.subscribe(ctx, sub, syms...)
		if err != nil {
			if err := ctx.Err(); err != nil {
				return err
			}

			log.Printf("Coinbase: Error: %+v", err)
			time.Sleep(x.retryAfter)
			continue
		}
		return nil
	}
}

func (x *Exchange) subscribe(ctx context.Context, sub exchange.Subscriber, syms ...string) error {
	ws, err := x.dial(ctx)
	if err != nil {
		return err
	}
	defer ws.Close()

	var ids = make([]string, len(syms))
	for i, s := range syms {
		ids[i] = symbol2coinbase(s)
	}

	req := coinbasepro.Message{
		Type: "subscribe",
		Channels: []coinbasepro.MessageChannel{
			coinbasepro.MessageChannel{Name: "level2", ProductIds: ids},
		},
	}

	log.Printf("Coinbase querying: %q", ids)
	if err := ws.WriteJSON(req); err != nil {
		return err
	}

	errCh := make(chan error)
	go func() {
		for {
			errCh <- x.handleJSON(ws, sub)
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

func (x *Exchange) handleJSON(ws *websocket.Conn, sub exchange.Subscriber) error {
	var msg coinbasepro.Message
	if err := ws.ReadJSON(&msg); err != nil {
		return err
	}

	var fn func(string, *obm.Update) error
	switch msg.Type {
	case "snapshot":
		fn = sub.OnSnapshot
	case "l2update":
		fn = sub.OnUpdate
	}

	if fn != nil {
		updates, err := newUpdates(&msg)
		if err != nil {
			return err
		}

		return fn(NAME, updates)
	}

	return nil
}

// NewUpdate returns a new obm.Update given a coinbasepro.Message
func newUpdates(msg *coinbasepro.Message) (*obm.Update, error) {
	var updates = obm.Update{
		Time:   msg.Time.Time(),
		Symbol: coinbase2symbol(msg.ProductID),
	}

	for _, bid := range msg.Bids {
		entry, err := obm.NewEntryFromStrings(bid.Price, bid.Size)
		if err != nil {
			return nil, err
		}
		updates.Bids = append(updates.Bids, entry)
	}

	for _, ask := range msg.Asks {
		entry, err := obm.NewEntryFromStrings(ask.Price, ask.Size)
		if err != nil {
			return nil, err
		}

		updates.Asks = append(updates.Asks, entry)
	}

	for _, change := range msg.Changes {
		entry, err := obm.NewEntryFromStrings(change.Price, change.Size)
		if err != nil {
			return nil, err
		}

		switch change.Side {
		case "buy":
			updates.Bids = append(updates.Bids, entry)
		case "sell":
			updates.Asks = append(updates.Asks, entry)
		}
	}

	return &updates, nil
}

func init() {
	exchange.Register(NAME, New())
}
