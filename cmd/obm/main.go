package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	flag "github.com/spf13/pflag"

	"github.com/ParadigmFoundation/go-obm/exchange"
	_ "github.com/ParadigmFoundation/go-obm/exchange/binance"
	_ "github.com/ParadigmFoundation/go-obm/exchange/coinbase"
	_ "github.com/ParadigmFoundation/go-obm/exchange/gemini"

	"github.com/ParadigmFoundation/go-obm/grpc"
	"github.com/ParadigmFoundation/go-obm/store/mem"
)

type xchange struct {
	name    string
	symbols []string
}

type config struct {
	bind      string
	exchanges []xchange
}

func parseArgs() (*config, error) {
	bind := flag.String("bind", "localhost:8000", "Webserver binding address")
	xz := flag.StringArray("exchange", nil, "Specify exchanges and symbols using <exchange>:<symbol[,symbol|...], you can use this flag many times to specify more than one exchange.\nExample:\n\t--exchange=coinbase:BTC-USD,ETH-USD --echange=binance:BTCUSDT")
	flag.Parse()

	cfg := &config{
		bind: *bind,
	}

	for _, x := range *xz {
		split := strings.Split(x, ":")
		if len(split) != 2 {
			return nil, fmt.Errorf("invalid format %s, use <echange>:<symbol[,symbol|...]", split)
		}

		xch := xchange{
			name:    split[0],
			symbols: strings.Split(split[1], ","),
		}
		if len(xch.symbols) == 0 || xch.symbols[0] == "" {
			return nil, fmt.Errorf("invalid format %s, no symbols defined, use <echange>:<symbol[,symbol|...]", split)
		}

		cfg.exchanges = append(cfg.exchanges, xch)
	}
	if len(cfg.exchanges) == 0 {
		return nil, fmt.Errorf("please specify at least one exchange")
	}

	return cfg, nil
}

func main() {
	cfg, err := parseArgs()
	if err != nil {
		log.Fatal(err)
	}

	store := mem.New()
	srv := grpc.NewServer(store)

	log.Printf("API Listening on %s", cfg.bind)
	go func() { _ = srv.Listen(cfg.bind) }()

	errCh := make(chan error)
	ctx := context.Background()
	for _, i := range cfg.exchanges {
		xch, err := exchange.Get(i.name)
		if err != nil {
			log.Fatalf("unknown exchange %s. Available exchanges: %q", i.name, exchange.ListNames())
		}

		syms := i.symbols
		go func() {
			err := xch.Subscribe(ctx, store, syms...)
			if err != nil {
				errCh <- err
			}
		}()
	}

	log.Fatal(<-errCh)
	srv.Stop()
}
