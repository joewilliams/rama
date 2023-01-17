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

func (t *TopK) GetIPs() map[netip.Addr]uint64 {
	output := map[netip.Addr]uint64{}
	for _, entry := range t.minHeap.nodes {
		output[entry.ip] = entry.count
	}

	return output
}

func (t *TopK) RankIPs() []netip.Addr {
	list := make([]netip.Addr, t.k)
	for i, entry := range t.minHeap.sort() {
		list[i] = entry.ip
	}

	return list
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

func (t *TopK) AddIP(ip netip.Addr) {
	var ipBytes []byte
	var fingerprint uint64

	idx, exists := t.minHeap.findByIP(ip)

	// if we've seen this IP before use the fingerprint and []byte we created previously
	if exists {
		ipBytes = t.minHeap.get(idx).data
		fingerprint = t.minHeap.get(idx).fingerprint
	} else {
		ipBytes = ip.AsSlice()
		fingerprint = t.xxhash(ipBytes)
	}

	maxCount := t.add(exists, ipBytes, fingerprint)

	if exists {
		t.minHeap.fix(idx, maxCount)
	} else {
		t.minHeap.add(node{
			count:       maxCount,
			ip:          ip,
			data:        ipBytes,
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

	for i := uint32(0); i < t.depth; i++ {
		binary.LittleEndian.PutUint32(bI, i)

		bucket := t.xxhash(append(data, bI...)) % t.width
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
