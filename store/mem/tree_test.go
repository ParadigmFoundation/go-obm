package mem

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gotest.tools/assert"
)

func TestCappedTree(t *testing.T) {
	keys := []float64{
		2,
		1.3,
		4,
		1.0,
		1.1,
		1.2,
		1.3,
		1.3,
		1.1,
		2.1,
		3.1,
	}

	t.Run("RemoveMin", func(t *testing.T) {
		tree := NewCappedTree(2, RemoveMin)
		for _, key := range keys {
			tree.Put(key, 0)
		}
		require.Equal(t, 2, tree.Size())
		assert.Equal(t, 3.1, tree.Keys()[0])
		assert.Equal(t, 4.0, tree.Keys()[1])
	})

	t.Run("RemoveMax", func(t *testing.T) {
		tree := NewCappedTree(2, RemoveMax)
		for _, key := range keys {
			tree.Put(key, 0)
		}
		require.Equal(t, 2, tree.Size())
		assert.Equal(t, 1.0, tree.Keys()[0])
		assert.Equal(t, 1.1, tree.Keys()[1])
	})

}
