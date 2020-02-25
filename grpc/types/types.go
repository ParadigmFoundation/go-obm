package types

//go:generate protoc --go_out=plugins=grpc:. types.proto

type OrderBookEntries []*OrderBookEntry

type OrderBookEntriesByPriceDesc OrderBookEntries

func (e OrderBookEntriesByPriceDesc) Len() int           { return len(e) }
func (e OrderBookEntriesByPriceDesc) Less(i, j int) bool { return e[i].Price > e[j].Price }
func (e OrderBookEntriesByPriceDesc) Swap(i, j int)      { e[i], e[j] = e[j], e[i] }

type OrderBookEntriesByPriceAsc OrderBookEntries

func (e OrderBookEntriesByPriceAsc) Len() int           { return len(e) }
func (e OrderBookEntriesByPriceAsc) Less(i, j int) bool { return e[i].Price < e[j].Price }
func (e OrderBookEntriesByPriceAsc) Swap(i, j int)      { e[i], e[j] = e[j], e[i] }
