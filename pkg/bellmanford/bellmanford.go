package bellmanford

import (
	"fmt"
	"math"
	"slices"
)

type Edge struct {
	src    string
	dst    string
	weight float64
}

type Graph struct {
	vertices []string
	edges    []Edge
}

func NewGraph() *Graph {
	return &Graph{
		// keep the list of vertices around so we dont need to generate them during path finding
		vertices: make([]string, 0),
		edges:    make([]Edge, 0),
	}
}

func (g *Graph) AddOrUpdateEdge(src string, dst string, weight float64) {
	for i, e := range g.edges {
		// update existing edge
		if e.src == src && e.dst == dst && e.weight != weight {
			g.edges = slices.Replace(g.edges, i, i, Edge{src, dst, weight})
			return
		}

		// duplicate, quit now
		if e.src == src && e.dst == dst && e.weight == weight {
			return
		}
	}

	// new edge, add it and update the vertices
	g.edges = append(g.edges, Edge{src, dst, weight})

	// if the src or dst is a new node update the vertices
	if !slices.Contains(g.vertices, src) {
		g.vertices = append(g.vertices, src)
	}

	if !slices.Contains(g.vertices, dst) {
		g.vertices = append(g.vertices, dst)
	}

}

func (g *Graph) BellmanFord(src string, dst string) ([]string, error) {
	if !slices.Contains(g.vertices, src) {
		return nil, fmt.Errorf("invalid source vertex: %v", src)
	}

	dist := make(map[string]float64, len(g.vertices))
	pred := make(map[string]string, len(g.vertices))

	for _, vert := range g.vertices {
		dist[vert] = math.Inf(1)
	}

	dist[src] = 0

	for range g.vertices {
		changed := false

		for _, e := range g.edges {
			if dist[e.src] != math.Inf(1) {
				if dist[e.src]+e.weight < dist[e.dst] {
					dist[e.dst] = dist[e.src] + e.weight
					pred[e.dst] = e.src
					changed = true
				}
			}
		}

		if !changed {
			break
		}
	}

	for _, e := range g.edges {
		if dist[e.src] != math.Inf(1) {
			if dist[e.src]+e.weight < dist[e.dst] {
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
