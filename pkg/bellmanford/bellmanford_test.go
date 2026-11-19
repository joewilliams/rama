package bellmanford

import (
	"fmt"
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

	fmt.Println(g)

	assert.NoError(t, err)
	assert.Equal(t, []string{"a", "b", "c", "e"}, path)

	g.AddOrUpdateEdge("a", "e", -1)
	path, err = g.BellmanFord("a", "e")

	assert.NoError(t, err)
	assert.Equal(t, []string{"a", "e"}, path)
}

func BenchmarkAll(b *testing.B) {
	g := NewGraph()
	for n := 0; n < b.N; n++ {
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
		g.AddOrUpdateEdge("b", "e", 3)
		g.AddOrUpdateEdge("b", "e", 3)
		g.AddOrUpdateEdge("b", "e", 3)
		g.AddOrUpdateEdge("b", "e", 3)
		g.AddOrUpdateEdge("b", "e", 3)
		g.AddOrUpdateEdge("b", "e", 3)
		g.BellmanFord("a", "e")
	}
}
