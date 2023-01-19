package rendezvous

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"net/netip"
	"time"

	"github.com/OneOfOne/xxhash"
)

const (
	// number of table entries * 100 == table size
	multiple = 100
)

type Table struct {
	members []netip.Addr
	table   []netip.Addr
	size    uint32
	key     uint64
}

func New(key uint64, members []netip.Addr) (Table, error) {
	return NewWithTableSize(key, uint32(len(members)*int(multiple)), members)
}

func NewWithTableSize(key uint64, size uint32, members []netip.Addr) (Table, error) {
	if len(members) < 1 {
		return Table{}, fmt.Errorf("too few members: %v", len(members))
	}

	if key == 0 {
		rand.Seed(time.Now().UnixNano())
		key = rand.Uint64()
	}

	table := Table{
		members: members,
		size:    size,
		key:     key,
	}

	table.generateTable()

	return table, nil
}

func (t *Table) Key() uint64 {
	return t.key
}

func (t *Table) Get(ip netip.Addr) netip.Addr {
	return t.table[t.xxhash(ip.AsSlice())&uint64(t.size-1)]
}

func (t *Table) Add(ip netip.Addr) {
	t.members = append(t.members, ip)
	t.generateTable()
}

func (t *Table) Delete(ip netip.Addr) {
	newMembers := []netip.Addr{}
	for _, entry := range t.members {
		if entry != ip {
			newMembers = append(newMembers, entry)
		}
	}
	t.members = newMembers
	t.generateTable()
}

func (t *Table) generateTable() {
	table := make([]netip.Addr, t.size)
	bI := make([]byte, 4)

	entrySlices := make([][]byte, len(t.members))
	for e, entry := range t.members {
		entrySlices[e] = entry.AsSlice()
	}

	for i := uint32(0); i < t.size; i++ {
		var highScore uint64
		var highEntry netip.Addr

		binary.LittleEndian.PutUint32(bI, i)

		for e, entry := range t.members {
			// hash the entry plus the table row index
			sum := t.xxhash(append(entrySlices[e], bI...))

			if sum > highScore {
				highScore = sum
				highEntry = entry
			}
		}

		table[i] = highEntry
	}

	t.table = table
}

func (t *Table) xxhash(data []byte) uint64 {
	return xxhash.Checksum64S(data, t.key)
}
