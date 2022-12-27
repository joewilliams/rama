package rama_rendezvous

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"net/netip"
	"time"

	"github.com/dchest/siphash"
	"github.com/twotwotwo/sorts/sortutil"
)

const (
	// number of table entries * 100 == table size
	multiple = 100
)

type Table struct {
	members []netip.Addr
	table   []netip.Addr
	size    int
	keyOne  int
	keyTwo  int
}

func New(keyOne int, keyTwo int, members []netip.Addr) (Table, error) {
	return NewWithTableSize(keyOne, keyTwo, len(members)*int(multiple), members)
}

func NewWithTableSize(keyOne int, keyTwo int, size int, members []netip.Addr) (Table, error) {
	if len(members) < 1 {
		return Table{}, fmt.Errorf("too few members: %v", len(members))
	}

	rand.Seed(time.Now().UnixNano())

	if keyOne == 0 {
		keyOne = rand.Int()
	}

	if keyTwo == 0 {
		keyTwo = rand.Int()
	}

	table := Table{
		members: members,
		size:    size,
		keyOne:  keyOne,
		keyTwo:  keyTwo,
	}

	table.generateTable()

	return table, nil
}

func (t *Table) GetKeys() (int, int) {
	return t.keyOne, t.keyTwo
}

func (t *Table) Get(ip netip.Addr) netip.Addr {
	sum := hash(t.keyOne, t.keyTwo, ip.AsSlice())
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

	for i := 0; i < t.size; i++ {
		rowEntries := map[uint64]netip.Addr{}
		rowKeys := make([]uint64, len(t.members))

		bI := make([]byte, 4)
		binary.LittleEndian.PutUint32(bI, uint32(i))

		for e, entry := range t.members {
			// hash the entry plus the table row index
			sum := hash(t.keyOne, t.keyTwo, append(entry.AsSlice(), bI...))
			rowEntries[sum] = entry
			rowKeys[e] = sum
		}

		sortutil.Uint64Slice(rowKeys).Sort()
		table[i] = rowEntries[rowKeys[0]]
	}

	t.table = table
}

func hash(seed int, key int, data []byte) uint64 {
	return siphash.Hash(uint64(seed), uint64(key), data)
}
