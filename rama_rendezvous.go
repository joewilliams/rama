package rama_rendezvous

import (
	"fmt"
	"net/netip"
	"sort"

	"github.com/dchest/siphash"
)

const (
	// number of table entries * 100 == table size
	multiple = 100
)

type Table struct {
	entries []netip.Addr
	table   []netip.Addr
	size    int
	keyOne  int
	keyTwo  int
}

func New(keyOne int, keyTwo int, entries []netip.Addr) (Table, error) {
	return NewWithTableSize(keyOne, keyTwo, len(entries)*int(multiple), entries)
}

func NewWithTableSize(keyOne int, keyTwo int, size int, entries []netip.Addr) (Table, error) {
	table := Table{
		entries: entries,
		size:    size,
		keyOne:  keyOne,
		keyTwo:  keyTwo,
	}

	table.generateTable()

	return table, nil
}

func (t *Table) Get(ip netip.Addr) netip.Addr {
	sum := hash(t.keyOne, t.keyTwo, ip.AsSlice())
	index := sum & uint64(t.size-1)
	return t.table[index]
}

func (t *Table) Add(ip netip.Addr) {
	t.entries = append(t.entries, ip)
	t.generateTable()
}

func (t *Table) Delete(ip netip.Addr) {
	newEntries := []netip.Addr{}
	for _, entry := range t.entries {
		if entry != ip {
			newEntries = append(newEntries, entry)
		}
	}
	t.entries = newEntries
	t.generateTable()
}

func (t *Table) generateTable() {
	table := make([]netip.Addr, t.size)

	for i := 0; i < t.size; i++ {
		rowEntries := map[uint64]netip.Addr{}
		rowKeys := make([]uint64, len(t.entries))
		for e, entry := range t.entries {
			// hash the entry plus the table row index
			sum := hash(t.keyOne, t.keyTwo, append(entry.AsSlice(), []byte(fmt.Sprint(i))...))
			rowEntries[sum] = entry
			rowKeys[e] = sum
		}
		sort.Slice(rowKeys, func(i, j int) bool { return rowKeys[i] < rowKeys[j] })
		table[i] = rowEntries[rowKeys[0]]
	}
	t.table = table
}

func hash(seed int, key int, data []byte) uint64 {
	return siphash.Hash(uint64(seed), uint64(key), data)
}
