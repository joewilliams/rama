package heavykeeper

import (
	"encoding/binary"
	"math"
	"math/rand"
	"net/netip"
	"time"

	"github.com/OneOfOne/xxhash"
)

type TopK struct {
	k       uint32
	width   uint64
	depth   uint32
	decay   float64
	seed    uint64
	buckets []nodes
	minHeap Heap
}

type node struct {
	ip          netip.Addr
	count       uint64
	fingerprint uint64
	slice       []byte
}

type nodes []node

func New(k uint32, width uint64, depth uint32, decay float64, seed uint64) TopK {
	rand.Seed(time.Now().UnixNano())

	buckets := make([]nodes, depth)
	for i := range buckets {
		buckets[i] = make(nodes, width)
	}

	topk := TopK{
		k:       k,
		width:   width,
		depth:   depth,
		decay:   decay,
		buckets: buckets,
		minHeap: newHeap(k),
		seed:    seed,
	}

	return topk
}

func (topk *TopK) Get() map[netip.Addr]uint64 {
	output := map[netip.Addr]uint64{}
	for _, entry := range topk.minHeap.nodes {
		output[entry.ip] = entry.count
	}

	return output
}

func (topk *TopK) Rank() []netip.Addr {
	list := make([]netip.Addr, topk.k)
	for i, entry := range topk.minHeap.sort() {
		list[i] = entry.ip
	}

	return list
}

func (topk *TopK) Add(ip netip.Addr) {
	var ipSlice []byte
	var ipFingerprint uint64

	idx, exists := topk.minHeap.find(ip)

	// if we've seen this IP before use the fingerprint and []byte we created previously
	if exists {
		ipSlice = topk.minHeap.get(idx).slice
		ipFingerprint = topk.minHeap.get(idx).fingerprint
	} else {
		ipSlice = ip.AsSlice()
		ipFingerprint = xxhash.Checksum64S(ipSlice, topk.seed)
	}

	maxCount := topk.add(exists, ipSlice, ipFingerprint)

	if exists {
		topk.minHeap.fix(idx, maxCount)
	} else {
		topk.minHeap.add(node{
			count:       maxCount,
			ip:          ip,
			slice:       ipSlice,
			fingerprint: ipFingerprint,
		})
	}
}

func (topk *TopK) add(exists bool, data []byte, fingerprint uint64) uint64 {
	bI := make([]byte, 4)
	min := topk.minHeap.min()
	var maxCount uint64

	for i := uint32(0); i < topk.depth; i++ {
		binary.LittleEndian.PutUint32(bI, i)

		bucket := xxhash.Checksum64S(append(data, bI...), topk.seed) % topk.width
		count := topk.buckets[i][bucket].count

		if count == 0 {
			topk.buckets[i][bucket].fingerprint = fingerprint
			topk.buckets[i][bucket].count++
			maxCount = max(maxCount, 1)
			continue
		}

		if topk.buckets[i][bucket].fingerprint == fingerprint {
			if exists || count <= min {
				topk.buckets[i][bucket].count++
				maxCount = max(maxCount, topk.buckets[i][bucket].count)
			}
			continue
		}

		if rand.Float64() < math.Pow(topk.decay, float64(count)) {
			topk.buckets[i][bucket].count--
			if topk.buckets[i][bucket].count == 0 {
				topk.buckets[i][bucket].fingerprint = fingerprint
				topk.buckets[i][bucket].count++
				maxCount = max(maxCount, 1)
			}
		}
	}

	return maxCount
}

func max(x, y uint64) uint64 {
	if x > y {
		return x
	}
	return y
}
