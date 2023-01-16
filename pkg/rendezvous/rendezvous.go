package rendezvous

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"net/netip"
	"time"

	"github.com/OneOfOne/xxhash"
	"golang.org/x/exp/slices"
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

	rand.Seed(time.Now().UnixNano())

	if key == 0 {
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
	return t.table[hash(t.key, ip.AsSlice())&uint64(t.size-1)]
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
	rowKeys := make([]uint64, len(t.members))
	bI := make([]byte, 4)

	entrySlices := make([][]byte, len(t.members))
	for e, entry := range t.members {
		entrySlices[e] = entry.AsSlice()
	}

	for i := uint32(0); i < t.size; i++ {
		rowEntries := map[uint64]netip.Addr{}
		binary.LittleEndian.PutUint32(bI, i)

		for e, entry := range t.members {
			// hash the entry plus the table row index
			sum := hash(t.key, append(entrySlices[e], bI...))
			rowEntries[sum] = entry
			rowKeys[e] = sum
		}

		slices.Sort(rowKeys)
		table[i] = rowEntries[rowKeys[0]]
	}

	t.table = table
}

func hash(key uint64, data []byte) uint64 {
	return xxhash.Checksum64S(data, uint64(key))
}
