package coinbase

import (
	"context"
	"sync"
	"testing"

	"github.com/gorilla/websocket"
	coinbasepro "github.com/preichenberger/go-coinbasepro/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ParadigmFoundation/go-obm"
	"github.com/ParadigmFoundation/go-obm/exchange/exchangetest"
	"github.com/ParadigmFoundation/go-obm/store/mem"
)

func assertRequestSymbols(done chan struct{}) exchangetest.WsFn {
	fn := func(t *testing.T, conn *websocket.Conn) {
		var req coinbasepro.Message
		err := conn.ReadJSON(&req)
		require.NoError(t, err)

		require.Len(t, req.Channels, 1)
		require.Equal(t, []string{"BTC-USD", "BTC-ETH"}, req.Channels[0].ProductIds)
		done <- struct{}{}
	}
	return fn
}

func TestCoinbaseRequest(t *testing.T) {
	done := make(chan struct{})

	// Start the fake coinbase server
	ws := exchangetest.NewWS(t, assertRequestSymbols(done))
	url, closer := ws.Start()
	defer closer()

	// start the exchange implementation
	x := New(WithURL(url))

	// Subscribe
	ctx := context.Background()
	sub := mem.New()
	go x.Subscribe(ctx, sub, "BTC/USD", "BTC/ETH") // nolint:errcheck
	<-done
}

func handleChanges(msgs chan interface{}, done chan struct{}) exchangetest.WsFn {
	fn := func(t *testing.T, conn *websocket.Conn) {
		// read the request but don't do anything with it
		require.NoError(t,
			conn.ReadJSON(&coinbasepro.Message{}),
		)

		for {
			msg, ok := <-msgs
			if !ok {
				break
			}
			_ = conn.WriteJSON(msg)
		}

		// finish
		require.NoError(t, conn.Close())
		done <- struct{}{}
	}
	return fn
}

// JsonMessage is the wire format, this is the json Coinbase sends us
type JsonMessage struct {
	Type      string     `json:"type"`
	ProductID string     `json:"product_id"`
	Asks      [][]string `json:"asks"`
	Bids      [][]string `json:"bids"`
	Changes   [][]string `json:"changes"`
}

func TestCoinbaseChanges(t *testing.T) {
	updates := make(chan *obm.Update)
	msgs := make(chan interface{})
	done := make(chan struct{})

	// Start the fake coinbase server
	ws := exchangetest.NewWS(t, handleChanges(msgs, done))
	url, closer := ws.Start()
	defer closer()

	// start the exchange implementation
	x := New(WithURL(url))

	// Subscribe
	ctx, cancel := context.WithCancel(context.Background())
	sub := exchangetest.NewFakeSubscriber(updates)
	go x.Subscribe(ctx, sub, "BTC/USD") // nolint:errcheck

	msgs <- &JsonMessage{
		Type:      "snapshot",
		ProductID: "BTC-USD",
		Bids: [][]string{
			{"1", "1"},
		},
		Asks: [][]string{
			{"2", "2"},
		},
		Changes: [][]string{
			{"buy", "3", "3"},
			{"sell", "3", "3"},
		},
	}

	up := <-updates
	assert.Equal(t, "BTC/USD", up.Symbol)
	assert.Len(t, up.Bids, 2)
	assert.Len(t, up.Asks, 2)

	cancel()
	close(msgs)
	<-done
}

func TestCoinbaseReconnection(t *testing.T) {
	connMutex := sync.Mutex{}
	connReady := make(chan struct{})
	var conn *websocket.Conn
	ws := exchangetest.NewWS(t, func(t *testing.T, c *websocket.Conn) {
		connMutex.Lock()
		conn = c
		connMutex.Unlock()

		connReady <- struct{}{}
	})
	url, closer := ws.Start()
	defer closer()

	x := New(WithURL(url), WithRetryAfter(0))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go x.Subscribe(ctx, exchangetest.NewFakeSubscriber(nil), "BTC/USD") // nolint:errcheck
	<-connReady

	connMutex.Lock()
	conn.Close()
	connMutex.Unlock()

	<-connReady
}
