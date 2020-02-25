package mem

import (
	"math/rand"
	"runtime"
	"testing"

	"github.com/ParadigmFoundation/go-obm"
	"github.com/ParadigmFoundation/go-obm/store"
	"github.com/ParadigmFoundation/go-obm/store/storetest"
	"github.com/stretchr/testify/require"
)

func TestMem(t *testing.T) {
	storetest.TestSuite(t, func(t *testing.T) (store.Store, func()) {
		return New(), func() {}
	})
}

func benchMem(b *testing.B, s store.Store, sym string) {
	// generate 3k updates
	for i := 0; i < 3000; i++ {
		update := &obm.Update{Symbol: sym,
			// Generate entries with prices ranging from 1 to 3000 and quantities ranging from 0 to 99
			Asks: obm.Entries{&obm.Entry{Price: float64(rand.Int31n(3000) + 1), Quantity: float64(rand.Int31n(300))}},
			Bids: obm.Entries{&obm.Entry{Price: float64(rand.Int31n(3000) + 1), Quantity: float64(rand.Int31n(300))}},
		}

		// update 5 exchanges
		_ = s.OnUpdate("mem1", update)
		_ = s.OnUpdate("mem2", update)
		_ = s.OnUpdate("mem3", update)
		_ = s.OnUpdate("mem4", update)
		_ = s.OnUpdate("mem5", update)
	}
}

func BenchmarkMem(b *testing.B) {
	var s store.Store

	sym := "BTC-USD"
	for i := 0; i < b.N; i++ {
		s = New()
		benchMem(b, s, sym)
	}

	ob, err := s.OrderBook("mem1", sym)
	require.NoError(b, err)
	b.Logf("mem -> Asks: %d, Bids: %d", len(ob.Asks), len(ob.Bids))

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	b.Logf("Total memory usage: %v/%v MiB (#%d)",
		m.Alloc/1024/1024, m.TotalAlloc/1024/1024, b.N,
	)
}
