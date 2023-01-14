package heavykeeper

import (
	"container/heap"
	"net/netip"

	"github.com/twotwotwo/sorts"
	"golang.org/x/exp/slices"
)

type Heap struct {
	nodes nodes
	k     uint32
}

func newHeap(k uint32) *Heap {
	h := nodes{}
	heap.Init(&h)
	return &Heap{nodes: h, k: k}
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

func (h *Heap) find(ip netip.Addr) (int, bool) {
	slices.SortFunc(h.nodes, func(i node, j node) bool {
		return i.ip.Less(j.ip)
	})

	return slices.BinarySearchFunc(h.nodes, node{ip: ip}, func(i node, j node) int {
		if i.ip == j.ip {
			return 0
		}

		if i.ip.Less(j.ip) {
			return -1
		}

		return 1
	})
}

func (h *Heap) sort() nodes {
	sorts.ByUint64(h.nodes)
	sorts.Flip(h.nodes)
	return h.nodes
}

// heap interface

func (n nodes) Len() int {
	return len(n)
}

func (n nodes) Less(i, j int) bool {
	return (n[i].count < n[j].count) || (n[i].count == n[j].count && n[i].ip.Less(n[j].ip))
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

func (n nodes) Key(i int) uint64 {
	return n[i].count
}
