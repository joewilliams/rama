package bellmanford

import (
	"fmt"
	"math"
	"slices"
)

type Edge struct {
	from   string
	to     string
	weight int64
}

type Graph struct {
	vertices []string
	edges    []Edge
}

func NewGraph() *Graph {
	return &Graph{
		vertices: make([]string, 0),
		edges:    make([]Edge, 0),
	}
}

func (g *Graph) AddOrUpdateEdge(from string, to string, weight int64) {
	g.edges = slices.DeleteFunc(g.edges, func(e Edge) bool {
		return e.from == from && e.to == to && e.weight != weight
	})

	g.edges = append(g.edges, Edge{from, to, weight})

	vertices := make([]string, 0, len(g.edges)*2)

	for _, e := range g.edges {
		vertices = append(vertices, e.from)
		vertices = append(vertices, e.to)
	}

	slices.Sort(vertices)
	g.vertices = slices.Compact(vertices)
}

func (g *Graph) BellmanFord(src string, dst string) ([]string, error) {
	if !slices.Contains(g.vertices, src) {
		return nil, fmt.Errorf("invalid source vertex: %v", src)
	}

	dist := make(map[string]int64, len(g.vertices))
	pred := make(map[string]string, len(g.vertices))

	for _, vert := range g.vertices {
		dist[vert] = int64(math.Inf(1))
		pred[vert] = ""
	}

	dist[src] = 0

	for range g.vertices {
		changed := false

		for _, e := range g.edges {
			if dist[e.from] != int64(math.Inf(1)) {
				if dist[e.from]+e.weight < dist[e.to] {
					dist[e.to] = dist[e.from] + e.weight
					pred[e.to] = e.from
					changed = true
				}
			}
		}

		if !changed {
			break
		}
	}

	for _, e := range g.edges {
		if dist[e.from] != int64(math.Inf(1)) {
			if dist[e.from]+e.weight < dist[e.to] {
				return nil, fmt.Errorf("negative cycle detected")
			}
		}
	}

	path := make([]string, 0, len(g.vertices))
	for v := dst; v != ""; v = pred[v] {
		path = append(path, v)
	}
	slices.Reverse(path)

	return path, nil
}
