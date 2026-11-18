package bellmanford

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBellmanFord(t *testing.T) {
	g := NewGraph()
	g.AddOrUpdateEdge("a", "b", -1)
	g.AddOrUpdateEdge("b", "c", 2)
	g.AddOrUpdateEdge("c", "d", 4)
	g.AddOrUpdateEdge("d", "e", -1)
	g.AddOrUpdateEdge("a", "c", 1)
	g.AddOrUpdateEdge("b", "d", 2)
	g.AddOrUpdateEdge("c", "e", -1)
	path, err := g.BellmanFord("a", "e")

	assert.NoError(t, err)
	assert.Equal(t, path, []string{"a", "b", "c", "e"})

	g.AddOrUpdateEdge("a", "e", -1)
	path, err = g.BellmanFord("a", "e")

	assert.NoError(t, err)
	assert.Equal(t, path, []string{"a", "e"})
}

func BenchmarkRun(b *testing.B) {
	g := NewGraph()
	g.AddOrUpdateEdge("a", "b", -1)
	g.AddOrUpdateEdge("b", "c", 2)
	g.AddOrUpdateEdge("c", "d", 4)
	g.AddOrUpdateEdge("d", "e", -1)
	g.AddOrUpdateEdge("a", "c", 1)
	g.AddOrUpdateEdge("b", "d", 2)
	g.AddOrUpdateEdge("c", "e", -1)
	g.AddOrUpdateEdge("c", "v", 1)
	g.AddOrUpdateEdge("a", "z", 1)
	g.AddOrUpdateEdge("b", "c", 2)
	g.AddOrUpdateEdge("z", "d", 2)
	g.AddOrUpdateEdge("c", "b", 2)
	g.AddOrUpdateEdge("c", "z", 3)
	g.AddOrUpdateEdge("c", "d", 3)
	g.AddOrUpdateEdge("b", "e", 3)

	for n := 0; n < b.N; n++ {
		g.BellmanFord("a", "e")
	}
}
