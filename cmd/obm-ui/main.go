package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/guptarohit/asciigraph"
	flag "github.com/spf13/pflag"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/ParadigmFoundation/go-obm/grpc"
	"github.com/ParadigmFoundation/go-obm/grpc/types"
)

func main() {
	addr := flag.String("addr", "localhost:8000", "Server address")
	xChng := flag.String("exchange", "coinbase", "Exchange [coinbase,binance,gemini]")
	symbol := flag.String("symbol", "BTC/USD", "Currency pair symbol")
	flag.Parse()

	ctx := context.Background()
	c, err := grpc.NewClient(ctx, *addr)
	if err != nil {
		log.Fatal(err)
	}

	for {
		time.After(100 * time.Millisecond)
		ob, err := c.OrderBook(ctx, &types.OrderBookRequest{Exchange: *xChng, Symbol: *symbol})
		if err != nil {
			log.Fatal(err)
		}

		render(ob, 1000)
	}
}

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func render(ob *types.OrderBookResponse, n int) {
	var bids []float64
	if l := len(ob.Bids); l > 0 {
		m := min(n, l)
		for _, bid := range ob.GetBids()[:m] {
			q := bid.GetQuantity()
			if len(bids) == 0 {
				bids = []float64{q}
			} else {
				bids = append(
					[]float64{(q + bids[0])}, bids...,
				)
			}
		}
	}

	var asks []float64
	if l := len(ob.Asks); l > 0 {
		m := min(n, l)
		for _, ask := range ob.GetAsks()[:m] {
			q := ask.GetQuantity()
			if len(asks) == 0 {
				asks = []float64{q}
			} else {
				asks = append(asks, q+asks[len(asks)-1])
			}
		}
	}

	var series = append(bids, asks...)
	if len(series) == 0 {
		return
	}

	w, h, _ := terminal.GetSize(0)
	caption := fmt.Sprintf("[%.2f - %.2f]", ob.GetBids()[0].Price, ob.GetAsks()[0].Price)
	caption = fmt.Sprintf("0%s%s", strings.Repeat(" ", (w/2)-(len(caption)+13)/2), caption)
	fmt.Printf("%s\n",
		asciigraph.Plot(series,
			asciigraph.Caption(caption),
			asciigraph.Height(h-3),
			asciigraph.Width(w-7),
		),
	)
}
