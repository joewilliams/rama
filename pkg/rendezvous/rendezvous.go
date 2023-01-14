package rendezvous

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"net/netip"
	"time"

	"github.com/OneOfOne/xxhash"
	"github.com/twotwotwo/sorts/sortutil"
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

func (t *Table) GetKey() uint64 {
	return t.key
}

func (t *Table) Get(ip netip.Addr) netip.Addr {
	sum := hash(t.key, ip.AsSlice())
	index := sum & uint64(t.size-1)
	return t.table[index]
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

	for i := uint32(0); i < t.size; i++ {
		rowEntries := map[uint64]netip.Addr{}
		rowKeys := make([]uint64, len(t.members))

		bI := make([]byte, 4)
		binary.LittleEndian.PutUint32(bI, i)

		for e, entry := range t.members {
			// hash the entry plus the table row index
			sum := hash(t.key, append(entry.AsSlice(), bI...))
			rowEntries[sum] = entry
			rowKeys[e] = sum
		}

		sortutil.Uint64Slice(rowKeys).Sort()
		table[i] = rowEntries[rowKeys[0]]
	}

	t.table = table
}

func hash(key uint64, data []byte) uint64 {
	return xxhash.Checksum64S(data, uint64(key))
}
