package mem

import (
	"github.com/emirpasic/gods/maps/treemap"
	"github.com/emirpasic/gods/utils"
)

type Remover int

const (
	RemoveMax Remover = iota
	RemoveMin
)

type CappedTree struct {
	*treemap.Map
	cap     int
	remover Remover
}

func NewCappedTree(cap int, r Remover) *CappedTree {
	return &CappedTree{
		Map:     treemap.NewWith(utils.Float64Comparator),
		cap:     cap,
		remover: r,
	}
}

func (t *CappedTree) Put(key, value float64) {
	if _, ok := t.Map.Get(key); ok {
		t.Map.Put(key, value)
		return
	}

	if t.Size() >= t.cap {
		switch t.remover {
		case RemoveMin:
			found, _ := t.Min()
			if key < found.(float64) {
				return
			}
			t.Map.Remove(found)
		case RemoveMax:
			found, _ := t.Max()
			if key > found.(float64) {
				return
			}
			t.Map.Remove(found)
		}
	}
	t.Map.Put(key, value)
}

func (t *CappedTree) Each(fn func(key, val float64)) {
	t.Map.Each(func(key, val interface{}) {
		fn(key.(float64), val.(float64))
	})
}
