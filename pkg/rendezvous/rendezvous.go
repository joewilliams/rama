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

type member struct {
	addr  netip.Addr
	bytes []byte
}

type Table struct {
	members []member
	table   []netip.Addr
	size    uint32
	key     uint64
}

func New(key uint64, membersList []netip.Addr) (Table, error) {
	return NewWithTableSize(key, uint32(len(membersList)*int(multiple)), membersList)
}

func NewWithTableSize(key uint64, size uint32, membersList []netip.Addr) (Table, error) {
	if len(membersList) < 1 {
		return Table{}, fmt.Errorf("too few members: %v", len(membersList))
	}

	if key == 0 {
		internalRand := rand.New(rand.NewSource(time.Now().UnixNano()))
		key = internalRand.Uint64()
	}

	members := make([]member, 0, len(membersList))
	for _, m := range membersList {
		members = append(members, member{addr: m, bytes: m.AsSlice()})
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

func (t *Table) Get(addr netip.Addr) netip.Addr {
	return t.table[t.xxhash(addr.AsSlice())&uint64(t.size-1)]
}

func (t *Table) Add(addr netip.Addr) {
	t.members = append(t.members, member{addr: addr, bytes: addr.AsSlice()})
	t.generateTable()
}

func (t *Table) Delete(addr netip.Addr) {
	newMembers := make([]member, 0, len(t.members))
	for _, member := range t.members {
		if member.addr != addr {
			newMembers = append(newMembers, member)
		}
	}
	t.members = newMembers
	t.generateTable()
}

func (t *Table) generateTable() {
	table := make([]netip.Addr, t.size)
	bI := make([]byte, 4)
	data := make([]byte, 0, 20) // 16+4 enough for v6 addr + bI

	for i := uint32(0); i < t.size; i++ {
		var highScore uint64
		var highMember netip.Addr

		binary.LittleEndian.PutUint32(bI, i)

		for _, member := range t.members {
			// hash the entry plus the table row index
			data = append(data, member.bytes...)
			data = append(data, bI...)
			sum := t.xxhash(data)
			data = data[:0] // clear it out before we use it again

			if sum > highScore {
				highScore = sum
				highMember = member.addr
			}
		}

		table[i] = highMember
	}

	t.table = table
}

func (t *Table) xxhash(data []byte) uint64 {
	return xxhash.Checksum64S(data, t.key)
}
