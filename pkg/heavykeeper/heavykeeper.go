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
	rand    *rand.Rand
}

type node struct {
	ip          netip.Addr
	count       uint64
	fingerprint uint64
	data        []byte
}

type nodes []node

func New(k uint32, width uint64, depth uint32, decay float64) TopK {
	rand.Seed(time.Now().UnixNano())
	return NewWtihSeed(k, width, depth, decay, rand.Uint64())
}

func NewWtihSeed(k uint32, width uint64, depth uint32, decay float64, seed uint64) TopK {
	internalRand := rand.New(rand.NewSource(time.Now().UnixNano()))

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
		rand:    internalRand,
	}

	return topk
}

func (topk *TopK) GetIPs() map[netip.Addr]uint64 {
	output := map[netip.Addr]uint64{}
	for _, entry := range topk.minHeap.nodes {
		output[entry.ip] = entry.count
	}

	return output
}

func (topk *TopK) RankIPs() []netip.Addr {
	list := make([]netip.Addr, topk.k)
	for i, entry := range topk.minHeap.sort() {
		list[i] = entry.ip
	}

	return list
}

func (topk *TopK) RankBytes() ([][]byte, []uint64) {
	listBytes := make([][]byte, topk.k)
	listCounts := make([]uint64, topk.k)
	for i, entry := range topk.minHeap.sort() {
		listBytes[i] = entry.data
		listCounts[i] = entry.count
	}

	return listBytes, listCounts
}

func (topk *TopK) AddIP(ip netip.Addr) {
	var ipBytes []byte
	var fingerprint uint64

	idx, exists := topk.minHeap.findByIP(ip)

	// if we've seen this IP before use the fingerprint and []byte we created previously
	if exists {
		ipBytes = topk.minHeap.get(idx).data
		fingerprint = topk.minHeap.get(idx).fingerprint
	} else {
		ipBytes = ip.AsSlice()
		fingerprint = topk.xxhash(ipBytes)
	}

	maxCount := topk.add(exists, ipBytes, fingerprint)

	if exists {
		topk.minHeap.fix(idx, maxCount)
	} else {
		topk.minHeap.add(node{
			count:       maxCount,
			ip:          ip,
			data:        ipBytes,
			fingerprint: fingerprint,
		})
	}
}

func (topk *TopK) AddBytes(data []byte) {
	var fingerprint uint64

	idx, exists := topk.minHeap.findByBytes(data)

	// if we've seen this IP before use the fingerprint we created previously
	if exists {
		fingerprint = topk.minHeap.get(idx).fingerprint
	} else {
		fingerprint = topk.xxhash(data)
	}

	maxCount := topk.add(exists, data, fingerprint)

	if exists {
		topk.minHeap.fix(idx, maxCount)
	} else {
		topk.minHeap.add(node{
			count:       maxCount,
			data:        data,
			fingerprint: fingerprint,
		})
	}
}

func (topk *TopK) add(exists bool, data []byte, fingerprint uint64) uint64 {
	bI := make([]byte, 4)
	min := topk.minHeap.min()
	var maxCount uint64

	for i := uint32(0); i < topk.depth; i++ {
		binary.LittleEndian.PutUint32(bI, i)

		bucket := topk.xxhash(append(data, bI...)) % topk.width
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

		if topk.rand.Float64() < math.Pow(topk.decay, float64(count)) {
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

func (topk *TopK) xxhash(data []byte) uint64 {
	return xxhash.Checksum64S(data, topk.seed)
}

func max(x, y uint64) uint64 {
	if x > y {
		return x
	}
	return y
}
