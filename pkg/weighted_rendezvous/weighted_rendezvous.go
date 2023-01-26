package weighted_rendezvous

import (
	"encoding/binary"
	"fmt"
	"math"
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
	addr   netip.Addr
	weight float64
	bytes  []byte
}

type Table struct {
	members []member
	size    uint32
	key     uint64
	table   []netip.Addr
}

func New(key uint64, membersMap map[netip.Addr]float64) (Table, error) {
	return NewWithTableSize(key, uint32(len(membersMap)*int(multiple)), membersMap)
}

func NewWithTableSize(key uint64, size uint32, membersMap map[netip.Addr]float64) (Table, error) {
	if len(membersMap) < 1 {
		return Table{}, fmt.Errorf("too few members: %v", len(membersMap))
	}

	if key == 0 {
		internalRand := rand.New(rand.NewSource(time.Now().UnixNano()))
		key = internalRand.Uint64()
	}

	var members []member
	for k, v := range membersMap {
		members = append(members, member{addr: k, weight: v, bytes: k.AsSlice()})
	}

	table := Table{
		key:     key,
		members: members,
		size:    size,
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

func (t *Table) Add(addr netip.Addr, weight float64) {
	t.members = append(t.members, member{addr: addr, weight: weight, bytes: addr.AsSlice()})
	t.generateTable()
}

func (t *Table) Delete(addr netip.Addr) {
	var newMembers []member
	for _, member := range t.members {
		if member.addr != addr {
			newMembers = append(newMembers, member)
		}
	}
	t.members = newMembers
	t.generateTable()
}

func (t *Table) Set(addr netip.Addr, weight float64) {
	for m, member := range t.members {
		if member.addr == addr {
			t.members[m].weight = weight
			break
		}
	}

	t.generateTable()
}

func (t *Table) generateTable() {
	bI := make([]byte, 4)
	table := make([]netip.Addr, t.size)

	for i := uint32(0); i < t.size; i++ {
		var highScore float64
		var highMember netip.Addr

		binary.LittleEndian.PutUint32(bI, i)

		for _, member := range t.members {
			// hash the entry plus the table row index
			sum := t.xxhash(append(member.bytes, bI...))

			score := sumToScore(sum, member.weight)

			if score > highScore {
				highScore = score
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

func sumToScore(sum uint64, weight float64) float64 {
	// this seems to work but what do i know i am just a dog at at computer
	// https://github.com/golang/go/issues/12290
	// we need to convert the uint64 sum to a uniformly random float64
	floatSum := float64(sum>>10) * float64(1.0/9007199254740992.0)
	return math.Abs((1.0 / -math.Log(float64(floatSum))) * weight)
}
