package rendezvous

import (
	"fmt"
	"net/netip"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	ips := []netip.Addr{
		netip.MustParseAddr("192.0.2.111"),
		netip.MustParseAddr("192.0.2.112"),
		netip.MustParseAddr("192.0.2.113"),
	}

	table, err := New(1234567812345678, ips)
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

	assert.Equal(t, 98, count0)
	assert.Equal(t, 105, count1)
	assert.Equal(t, 97, count2)

	assert.Equal(t, len(table.table), count0+count1+count2)

	want := map[string]string{
		"192.0.2.1": "192.0.2.112",
		"192.0.2.2": "192.0.2.113",
		"192.0.2.3": "192.0.2.111",
		"192.0.2.4": "192.0.2.112",
		"192.0.2.5": "192.0.2.112",
	}

	for k, v := range want {
		//fmt.Printf("%v / %v\n", k, v)
		ip := table.Get(netip.MustParseAddr(k))
		assert.Equal(t, v, ip.String())
	}
}

func TestDelete(t *testing.T) {
	toDelete := netip.MustParseAddr("192.0.2.1")

	ips := []netip.Addr{
		netip.MustParseAddr("192.0.2.1"),
		netip.MustParseAddr("192.0.2.2"),
		netip.MustParseAddr("192.0.2.3"),
		netip.MustParseAddr("192.0.2.4"),
		netip.MustParseAddr("192.0.2.5"),
		netip.MustParseAddr("2001:0db8:85a3:1:1:8a2e:0370:7334"),
	}

	table, err := New(1234567812345678, ips)
	assert.Nil(t, err)

	want := map[string]netip.Addr{}

	for i := 0; i < 22; i++ {
		stringIP := fmt.Sprintf("192.0.2.%v", i)
		ip := table.Get(netip.MustParseAddr(stringIP))
		want[stringIP] = ip
	}

	// deleting 192.0.2.1 should not affect the others
	table.Delete(toDelete)

	count0 := 0
	count1 := 0
	count2 := 0
	count3 := 0
	count4 := 0
	count5 := 0

	for _, ip := range table.table {
		assert.NotEqual(t, ip, toDelete)
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
	assert.Equal(t, 133, count1)
	assert.Equal(t, 104, count2)
	assert.Equal(t, 113, count3)
	assert.Equal(t, 121, count4)
	assert.Equal(t, 129, count5)

	for k, v := range want {
		//fmt.Printf("%v / %v\n", k, v)
		ip := table.Get(netip.MustParseAddr(k))

		// we removed 192.0.2.1 so previous mappings for that ip should change
		if v.String() == toDelete.String() {
			assert.NotEqual(t, toDelete.String(), ip.String())
		} else {
			// the other mappings should not change
			assert.Equal(t, v.String(), ip.String())
		}
	}
}

func TestAdd(t *testing.T) {
	ips := []netip.Addr{
		netip.MustParseAddr("192.0.2.1"),
		netip.MustParseAddr("192.0.2.2"),
		netip.MustParseAddr("192.0.2.3"),
	}

	table, err := New(1234567812345678, ips)
	assert.Nil(t, err)

	// adding 192.0.2.4 should affect everything
	newMember := netip.MustParseAddr("2001:0db8:85a3:1:1:8a2e:0370:7334")
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

	assert.Equal(t, 73, count0)
	assert.Equal(t, 80, count1)
	assert.Equal(t, 71, count2)
	assert.Equal(t, 76, count3)
}

func TestGetKeys(t *testing.T) {
	ips := []netip.Addr{
		netip.MustParseAddr("192.0.2.1"),
		netip.MustParseAddr("192.0.2.2"),
		netip.MustParseAddr("192.0.2.3"),
	}

	table, err := New(9999, ips)
	assert.Nil(t, err)

	key := table.Key()
	assert.Equal(t, uint64(9999), key)
}

func TestBadNew(t *testing.T) {
	_, err := New(0, []netip.Addr{})
	assert.NotNil(t, err)
}

func BenchmarkGenerateOneEntry(b *testing.B) {
	ips := []netip.Addr{netip.MustParseAddr("192.0.2.1")}

	for n := 0; n < b.N; n++ {
		_, err := New(1234, ips)
		assert.Nil(b, err)
	}
}

func BenchmarkGenerateTenEntries(b *testing.B) {
	ips := []netip.Addr{}
	for i := 0; i < 10; i++ {
		ips = append(ips, netip.MustParseAddr(fmt.Sprintf("192.0.2.%v", i)))
	}

	for n := 0; n < b.N; n++ {
		_, err := New(1234, ips)
		assert.Nil(b, err)
	}
}

func BenchmarkGenerate1kEntries(b *testing.B) {
	ips := []netip.Addr{}
	for i := 0; i < 250; i++ {
		ips = append(ips, netip.MustParseAddr(fmt.Sprintf("192.0.1.%v", i)))
		ips = append(ips, netip.MustParseAddr(fmt.Sprintf("192.0.2.%v", i)))
		ips = append(ips, netip.MustParseAddr(fmt.Sprintf("192.0.3.%v", i)))
		ips = append(ips, netip.MustParseAddr(fmt.Sprintf("192.0.4.%v", i)))
	}

	for n := 0; n < b.N; n++ {
		New(1234, ips)
	}
}

func BenchmarkGenerateLookup(b *testing.B) {
	ips := []netip.Addr{
		netip.MustParseAddr("192.0.2.1"),
		netip.MustParseAddr("192.0.2.2"),
		netip.MustParseAddr("192.0.2.3"),
	}

	table, _ := New(1234, ips)

	lookupIP := netip.MustParseAddr("192.0.2.4")

	for n := 0; n < b.N; n++ {
		table.Get(lookupIP)
	}
}
