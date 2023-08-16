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
	addr        netip.Addr
	count       uint64
	fingerprint uint64
	data        []byte
}

type nodes []node

func New(k uint32, width uint64, depth uint32, decay float64) TopK {
	return NewWtihSeed(k, width, depth, decay, 0)
}

func NewWtihSeed(k uint32, width uint64, depth uint32, decay float64, seed uint64) TopK {
	internalRand := rand.New(rand.NewSource(time.Now().UnixNano()))

	if seed == 0 {
		seed = internalRand.Uint64()
	}

	buckets := make([]nodes, depth)
	for i := range buckets {
		buckets[i] = make(nodes, width)
	}

	t := TopK{
		k:       k,
		width:   width,
		depth:   depth,
		decay:   decay,
		buckets: buckets,
		minHeap: newHeap(k),
		seed:    seed,
		rand:    internalRand,
	}

	return t
}

func (t *TopK) GetAddrs() map[netip.Addr]uint64 {
	output := map[netip.Addr]uint64{}
	for _, entry := range t.minHeap.nodes {
		output[entry.addr] = entry.count
	}

	return output
}

func (t *TopK) GetBytes() map[string]uint64 {
	// we can't have nice things so use strings as keys
	output := map[string]uint64{}
	for _, entry := range t.minHeap.nodes {
		output[string(entry.data)] = entry.count
	}

	return output
}

func (t *TopK) RankAddrs() ([]netip.Addr, []uint64) {
	listAddrs := make([]netip.Addr, t.k)
	listCounts := make([]uint64, t.k)
	for i, entry := range t.minHeap.sort() {
		listAddrs[i] = entry.addr
		listCounts[i] = entry.count
	}

	return listAddrs, listCounts
}

func (t *TopK) RankBytes() ([][]byte, []uint64) {
	listBytes := make([][]byte, t.k)
	listCounts := make([]uint64, t.k)
	for i, entry := range t.minHeap.sort() {
		listBytes[i] = entry.data
		listCounts[i] = entry.count
	}

	return listBytes, listCounts
}

func (t *TopK) AddAddr(addr netip.Addr) {
	var addrBytes []byte
	var fingerprint uint64

	idx, exists := t.minHeap.findByAddr(addr)

	// if we've seen this IP before use the fingerprint and []byte we created previously
	if exists {
		addrBytes = t.minHeap.get(idx).data
		fingerprint = t.minHeap.get(idx).fingerprint
	} else {
		addrBytes = addr.AsSlice()
		fingerprint = t.xxhash(addrBytes)
	}

	maxCount := t.add(exists, addrBytes, fingerprint)

	if exists {
		t.minHeap.fix(idx, maxCount)
	} else {
		t.minHeap.add(node{
			count:       maxCount,
			addr:        addr,
			data:        addrBytes,
			fingerprint: fingerprint,
		})
	}
}

func (t *TopK) AddBytes(data []byte) {
	var fingerprint uint64

	idx, exists := t.minHeap.findByBytes(data)

	// if we've seen this IP before use the fingerprint we created previously
	if exists {
		fingerprint = t.minHeap.get(idx).fingerprint
	} else {
		fingerprint = t.xxhash(data)
	}

	maxCount := t.add(exists, data, fingerprint)

	if exists {
		t.minHeap.fix(idx, maxCount)
	} else {
		t.minHeap.add(node{
			count:       maxCount,
			data:        data,
			fingerprint: fingerprint,
		})
	}
}

func (t *TopK) add(exists bool, data []byte, fingerprint uint64) uint64 {
	bI := make([]byte, 4)
	min := t.minHeap.min()
	var maxCount uint64
	dataX := make([]byte, 0, 20) // 16+4 enough for v6 addr + bI

	for i := uint32(0); i < t.depth; i++ {
		binary.LittleEndian.PutUint32(bI, i)

		dataX = append(dataX, data...)
		dataX = append(dataX, bI...)
		bucket := t.xxhash(dataX) % t.width
		dataX = dataX[:0] // clear it out before we use it again
		count := t.buckets[i][bucket].count

		if count == 0 {
			t.buckets[i][bucket].fingerprint = fingerprint
			t.buckets[i][bucket].count++
			maxCount = max(maxCount, 1)
			continue
		}

		if t.buckets[i][bucket].fingerprint == fingerprint {
			if exists || count <= min {
				t.buckets[i][bucket].count++
				maxCount = max(maxCount, t.buckets[i][bucket].count)
			}
			continue
		}

		if t.rand.Float64() < math.Pow(t.decay, float64(count)) {
			t.buckets[i][bucket].count--
			if t.buckets[i][bucket].count == 0 {
				t.buckets[i][bucket].fingerprint = fingerprint
				t.buckets[i][bucket].count++
				maxCount = max(maxCount, 1)
			}
		}
	}

	return maxCount
}

func (t *TopK) xxhash(data []byte) uint64 {
	return xxhash.Checksum64S(data, t.seed)
}

func max(x, y uint64) uint64 {
	if x > y {
		return x
	}
	return y
}
