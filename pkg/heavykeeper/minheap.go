package heavykeeper

import (
	"bytes"
	"container/heap"
	"net/netip"

	"golang.org/x/exp/slices"
)

type Heap struct {
	nodes nodes
	k     uint32
}

func newHeap(k uint32) Heap {
	h := nodes{}
	heap.Init(&h)
	return Heap{nodes: h, k: k}
}

func (h *Heap) add(val node) {
	if h.k > uint32(len(h.nodes)) {
		heap.Push(&h.nodes, val)
	} else if val.count > h.nodes[0].count {
		heap.Push(&h.nodes, val)
		heap.Pop(&h.nodes)
	}
}

func (h *Heap) fix(idx int, count uint64) {
	h.nodes[idx].count = count
	heap.Fix(&h.nodes, idx)
}

func (h *Heap) min() uint64 {
	if len(h.nodes) == 0 {
		return 0
	}
	return h.nodes[0].count
}

func (h *Heap) get(i int) node {
	return h.nodes[i]
}

func (h *Heap) findByAddr(addr netip.Addr) (int, bool) {
	for i := range h.nodes {
		if h.nodes[i].addr == addr {
			return i, true
		}
	}
	return 0, false
}

func (h *Heap) findByBytes(data []byte) (int, bool) {
	for i := range h.nodes {
		if bytes.Equal(h.nodes[i].data, data) {
			return i, true
		}
	}
	return 0, false
}

func (h *Heap) sort() nodes {
	slices.SortFunc(h.nodes, func(a node, b node) bool {
		return a.count > b.count // using > to get reverse order
	})
	return h.nodes
}

// heap interface

func (n nodes) Len() int {
	return len(n)
}

func (n nodes) Less(i, j int) bool {
	// if we have IPs use those
	if n[i].addr.IsValid() && n[j].addr.IsValid() {
		return (n[i].count < n[j].count) || (n[i].count == n[j].count && n[i].addr.Less(n[j].addr))
	}

	// if we don't use the bytes we have
	return (n[i].count < n[j].count) || (n[i].count == n[j].count && bytes.Compare(n[i].data, n[j].data) < 0)
}

func (n nodes) Swap(i, j int) {
	n[i], n[j] = n[j], n[i]
}

func (n *nodes) Push(val interface{}) {
	*n = append(*n, val.(node))
}

func (n *nodes) Pop() interface{} {
	var val node
	val, *n = (*n)[len((*n))-1], (*n)[:len((*n))-1]
	return val
}
