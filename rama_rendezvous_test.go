package rama_rendezvous

import (
	"fmt"
	"net/netip"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	ips := []netip.Addr{
		netip.MustParseAddr("192.168.1.1"),
		netip.MustParseAddr("192.168.1.2"),
		netip.MustParseAddr("192.168.1.3"),
	}

	table, err := New(1234567812345678, 1234567812345678, ips)
	assert.Nil(t, err)

	assert.Equal(t, len(ips)*multiple, len(table.table))

	count0 := 0
	count1 := 0
	count2 := 0

	for _, ip := range table.table {
		//fmt.Println(ip.String())

		assert.True(t, ip.IsValid())

		if ip == ips[0] {
			count0++
		}

		if ip == ips[1] {
			count1++
		}

		if ip == ips[2] {
			count2++
		}
	}

	assert.Equal(t, 116, count0)
	assert.Equal(t, 91, count1)
	assert.Equal(t, 93, count2)

	assert.Equal(t, len(table.table), count0+count1+count2)

	testMap := map[string]string{
		"192.168.200.1":  "192.168.1.1",
		"192.168.200.2":  "192.168.1.1",
		"192.168.200.3":  "192.168.1.1",
		"192.168.200.4":  "192.168.1.3",
		"192.168.200.5":  "192.168.1.1",
		"192.168.200.6":  "192.168.1.2",
		"192.168.200.7":  "192.168.1.2",
		"192.168.200.8":  "192.168.1.1",
		"192.168.200.9":  "192.168.1.3",
		"192.168.200.10": "192.168.1.3",
		"192.168.200.11": "192.168.1.1",
		"192.168.200.12": "192.168.1.3",
		"192.168.200.13": "192.168.1.2",
		"192.168.200.14": "192.168.1.2",
		"192.168.200.15": "192.168.1.3",
		"192.168.200.16": "192.168.1.1",
		"192.168.200.17": "192.168.1.2",
		"192.168.200.18": "192.168.1.1",
		"192.168.200.19": "192.168.1.1",
		"192.168.200.20": "192.168.1.1",
		"192.168.200.21": "192.168.1.1",
		"192.168.200.22": "192.168.1.2",
	}

	for k, v := range testMap {
		//fmt.Printf("%v / %v\n", k, v)
		ip := table.Get(netip.MustParseAddr(k))
		assert.Equal(t, v, ip.String())
	}
}

func TestDelete(t *testing.T) {
	ips := []netip.Addr{
		netip.MustParseAddr("192.168.1.1"),
		netip.MustParseAddr("192.168.1.2"),
		netip.MustParseAddr("192.168.1.3"),
		netip.MustParseAddr("192.168.1.4"),
		netip.MustParseAddr("192.168.1.5"),
		netip.MustParseAddr("192.168.1.6"),
	}

	table, err := New(1234567812345678, 1234567812345678, ips)
	assert.Nil(t, err)

	testMap := map[string]netip.Addr{
		"192.168.200.1":  netip.Addr{},
		"192.168.200.2":  netip.Addr{},
		"192.168.200.3":  netip.Addr{},
		"192.168.200.4":  netip.Addr{},
		"192.168.200.5":  netip.Addr{},
		"192.168.200.6":  netip.Addr{},
		"192.168.200.7":  netip.Addr{},
		"192.168.200.8":  netip.Addr{},
		"192.168.200.9":  netip.Addr{},
		"192.168.200.10": netip.Addr{},
		"192.168.200.11": netip.Addr{},
		"192.168.200.12": netip.Addr{},
		"192.168.200.13": netip.Addr{},
		"192.168.200.14": netip.Addr{},
		"192.168.200.15": netip.Addr{},
		"192.168.200.16": netip.Addr{},
		"192.168.200.17": netip.Addr{},
		"192.168.200.18": netip.Addr{},
		"192.168.200.19": netip.Addr{},
		"192.168.200.20": netip.Addr{},
		"192.168.200.21": netip.Addr{},
		"192.168.200.22": netip.Addr{},
	}

	for k := range testMap {
		ip := table.Get(netip.MustParseAddr(k))
		testMap[k] = ip
	}

	// deleting 192.168.1.1 should not affect the others
	table.Delete(netip.MustParseAddr("192.168.1.1"))

	count0 := 0
	count1 := 0
	count2 := 0
	count3 := 0
	count4 := 0
	count5 := 0

	for _, ip := range table.table {
		assert.NotEqual(t, ip, netip.MustParseAddr("192.168.1.1"))
		assert.True(t, ip.IsValid())

		if ip == ips[0] {
			count0++
		}

		if ip == ips[1] {
			count1++
		}

		if ip == ips[2] {
			count2++
		}

		if ip == ips[3] {
			count3++
		}

		if ip == ips[4] {
			count4++
		}

		if ip == ips[5] {
			count5++
		}
	}

	assert.Equal(t, 0, count0)
	assert.Equal(t, 111, count1)
	assert.Equal(t, 116, count2)
	assert.Equal(t, 119, count3)
	assert.Equal(t, 134, count4)
	assert.Equal(t, 120, count5)

	for k, v := range testMap {
		//fmt.Printf("%v / %v\n", k, v)
		ip := table.Get(netip.MustParseAddr(k))

		// we removed 192.168.1.1 so previous mappings for that ip should change
		if v.String() == "192.168.1.1" {
			assert.NotEqual(t, "192.168.1.1", ip.String())
		} else {
			// the other mappings should not change
			assert.Equal(t, v.String(), ip.String())
		}
	}
}

func TestAdd(t *testing.T) {
	ips := []netip.Addr{
		netip.MustParseAddr("192.168.1.1"),
		netip.MustParseAddr("192.168.1.2"),
		netip.MustParseAddr("192.168.1.3"),
	}

	table, err := New(1234567812345678, 1234567812345678, ips)
	assert.Nil(t, err)

	// adding 192.168.1.4 should affect everything
	newMember := netip.MustParseAddr("192.168.1.4")
	table.Add(newMember)

	count0 := 0
	count1 := 0
	count2 := 0
	count3 := 0

	for _, ip := range table.table {
		//fmt.Println(ip.String())

		assert.True(t, ip.IsValid())

		if ip == ips[0] {
			count0++
		}

		if ip == ips[1] {
			count1++
		}

		if ip == ips[2] {
			count2++
		}

		if ip == newMember {
			count3++
		}
	}

	assert.Equal(t, 78, count0)
	assert.Equal(t, 69, count1)
	assert.Equal(t, 71, count2)
	assert.Equal(t, 82, count3)
}

func TestGetKeys(t *testing.T) {
	ips := []netip.Addr{
		netip.MustParseAddr("192.168.1.1"),
		netip.MustParseAddr("192.168.1.2"),
		netip.MustParseAddr("192.168.1.3"),
	}

	table, err := New(9999, 1234, ips)
	assert.Nil(t, err)

	keyOne, keyTwo := table.GetKeys()
	assert.Equal(t, 9999, keyOne)
	assert.Equal(t, 1234, keyTwo)
}

func TestBadNew(t *testing.T) {
	_, err := New(0, 0, []netip.Addr{})
	assert.NotNil(t, err)
}

func BenchmarkGenerateThreeEntries(b *testing.B) {
	ips := []netip.Addr{
		netip.MustParseAddr("192.168.1.1"),
		netip.MustParseAddr("192.168.1.2"),
		netip.MustParseAddr("192.168.1.3"),
	}

	for n := 0; n < b.N; n++ {
		New(1234, 9876, ips)
	}
}

func BenchmarkGenerate1kEntries(b *testing.B) {
	ips := []netip.Addr{}

	for i := 0; i < 250; i++ {
		ips = append(ips, netip.MustParseAddr(fmt.Sprintf("192.168.1.%v", i)))
		ips = append(ips, netip.MustParseAddr(fmt.Sprintf("192.168.2.%v", i)))
		ips = append(ips, netip.MustParseAddr(fmt.Sprintf("192.168.3.%v", i)))
		ips = append(ips, netip.MustParseAddr(fmt.Sprintf("192.168.4.%v", i)))
	}

	for n := 0; n < b.N; n++ {
		New(1234, 9876, ips)
	}
}
