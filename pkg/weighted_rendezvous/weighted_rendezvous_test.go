package weighted_rendezvous

import (
	"fmt"
	"net/netip"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	ips := map[netip.Addr]float64{
		netip.MustParseAddr("192.0.2.111"): 10,
		netip.MustParseAddr("192.0.2.112"): 20,
		netip.MustParseAddr("192.0.2.113"): 70,
	}

	table, err := New(1234567812345678, ips)
	assert.Nil(t, err)

	counts := map[netip.Addr]float64{
		netip.MustParseAddr("192.0.2.111"): 0,
		netip.MustParseAddr("192.0.2.112"): 0,
		netip.MustParseAddr("192.0.2.113"): 0,
	}

	for i := 0; i <= 255; i++ {
		for j := 0; j <= 255; j++ {
			lookup := netip.MustParseAddr(fmt.Sprintf("192.0.%v.%v", i, j))
			ip := table.Get(lookup)

			for k := range counts {
				if k == ip {
					counts[k]++
				}
			}
		}
	}

	total := float64(0)

	for k, v := range counts {
		percent := float64(v) / float64(65536)
		total = total + percent

		//fmt.Printf("%v : %v : %v\n", k, v, percent)

		// make sure the distribution is within +/- a few percent
		if k == netip.MustParseAddr("192.0.2.111") {
			assert.GreaterOrEqual(t, percent, 0.05)
			assert.LessOrEqual(t, percent, 0.15)
		}

		if k == netip.MustParseAddr("192.0.2.112") {
			assert.GreaterOrEqual(t, percent, 0.10)
			assert.LessOrEqual(t, percent, 0.25)
		}

		if k == netip.MustParseAddr("192.0.2.113") {
			assert.GreaterOrEqual(t, percent, 0.65)
			assert.LessOrEqual(t, percent, 0.80)
		}
	}

	// it should always equal 1
	assert.Equal(t, total, float64(1))

	toDelete := netip.MustParseAddr("192.0.2.111")
	want := map[string]netip.Addr{}

	for i := 0; i < 22; i++ {
		stringIP := fmt.Sprintf("192.0.2.%v", i)
		ip := table.Get(netip.MustParseAddr(stringIP))
		want[stringIP] = ip
	}

	// deleting 192.0.2.111 should not affect the others
	table.Delete(toDelete)

	for _, ip := range table.table {
		assert.NotEqual(t, ip, toDelete)
		assert.True(t, ip.IsValid())
	}

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

	newMember := netip.MustParseAddr("2001:0db8:85a3:1:1:8a2e:0370:7334")
	table.Add(newMember, 30)

	countsAdd := map[netip.Addr]float64{
		newMember:                          0,
		netip.MustParseAddr("192.0.2.112"): 0,
		netip.MustParseAddr("192.0.2.113"): 0,
	}

	for i := 0; i <= 255; i++ {
		for j := 0; j <= 255; j++ {
			lookup := netip.MustParseAddr(fmt.Sprintf("192.0.%v.%v", i, j))
			ip := table.Get(lookup)

			for k := range countsAdd {
				if k == ip {
					countsAdd[k]++
				}
			}
		}
	}

	totalAdd := float64(0)

	for k, v := range countsAdd {
		percent := float64(v) / float64(65536)
		totalAdd = totalAdd + percent

		//fmt.Printf("%v : %v : %v\n", k, v, percent)

		// make sure the distribution is within +/- a few percent
		if k == newMember {
			assert.GreaterOrEqual(t, percent, 0.15)
			assert.LessOrEqual(t, percent, 0.35)
		}

		if k == netip.MustParseAddr("192.0.2.112") {
			assert.GreaterOrEqual(t, percent, 0.10)
			assert.LessOrEqual(t, percent, 0.25)
		}

		if k == netip.MustParseAddr("192.0.2.113") {
			assert.GreaterOrEqual(t, percent, 0.65)
			assert.LessOrEqual(t, percent, 0.80)
		}
	}

	// it should always equal 1
	assert.Equal(t, totalAdd, float64(1))

	table.Set(netip.MustParseAddr("192.0.2.112"), 40)

	countsSet := map[netip.Addr]float64{
		newMember:                          0,
		netip.MustParseAddr("192.0.2.112"): 0,
		netip.MustParseAddr("192.0.2.113"): 0,
	}

	for i := 0; i <= 255; i++ {
		for j := 0; j <= 255; j++ {
			lookup := netip.MustParseAddr(fmt.Sprintf("192.0.%v.%v", i, j))
			ip := table.Get(lookup)

			for k := range countsSet {
				if k == ip {
					countsSet[k]++
				}
			}
		}
	}

	totalSet := float64(0)

	for k, v := range countsSet {
		percent := float64(v) / float64(65536)
		totalSet = totalSet + percent

		//fmt.Printf("%v : %v : %v\n", k, v, percent)

		// make sure the distribution is within +/- a few percent
		if k == newMember {
			assert.GreaterOrEqual(t, percent, 0.15)
			assert.LessOrEqual(t, percent, 0.35)
		}

		if k == netip.MustParseAddr("192.0.2.112") {
			assert.GreaterOrEqual(t, percent, 0.10)
			assert.LessOrEqual(t, percent, 0.20)
		}

		if k == netip.MustParseAddr("192.0.2.113") {
			assert.GreaterOrEqual(t, percent, 0.60)
			assert.LessOrEqual(t, percent, 0.80)
		}
	}

	// it should always equal 1
	assert.Equal(t, totalAdd, float64(1))
}

func TestGetKeys(t *testing.T) {
	ips := map[netip.Addr]float64{
		netip.MustParseAddr("192.0.2.1"): 0.1,
		netip.MustParseAddr("192.0.2.2"): 0.5,
		netip.MustParseAddr("192.0.2.3"): 0.4,
	}

	table, err := New(9999, ips)
	assert.Nil(t, err)

	key := table.Key()
	assert.Equal(t, uint64(9999), key)
}

func TestBadNew(t *testing.T) {
	_, err := New(0, map[netip.Addr]float64{})
	assert.NotNil(t, err)
}

func BenchmarkGenerateOneEntry(b *testing.B) {
	ips := map[netip.Addr]float64{netip.MustParseAddr("192.0.2.1"): 10}

	for n := 0; n < b.N; n++ {
		_, err := New(1234, ips)
		assert.Nil(b, err)
	}
}

func BenchmarkGenerateTenEntries(b *testing.B) {
	ips := map[netip.Addr]float64{}
	for i := 0; i < 10; i++ {
		ips[netip.MustParseAddr(fmt.Sprintf("192.0.2.%v", i))] = 10
	}

	for n := 0; n < b.N; n++ {
		_, err := New(1234, ips)
		assert.Nil(b, err)
	}
}

func BenchmarkGenerate1kEntries(b *testing.B) {
	ips := map[netip.Addr]float64{}
	for i := 0; i < 250; i++ {
		ips[netip.MustParseAddr(fmt.Sprintf("192.0.1.%v", i))] = 10
		ips[netip.MustParseAddr(fmt.Sprintf("192.0.2.%v", i))] = 10
		ips[netip.MustParseAddr(fmt.Sprintf("192.0.3.%v", i))] = 10
		ips[netip.MustParseAddr(fmt.Sprintf("192.0.4.%v", i))] = 10
	}

	for n := 0; n < b.N; n++ {
		New(1234, ips)
	}
}

func BenchmarkGenerateLookup(b *testing.B) {
	ips := map[netip.Addr]float64{
		netip.MustParseAddr("192.0.2.1"): 10,
		netip.MustParseAddr("192.0.2.2"): 20,
		netip.MustParseAddr("192.0.2.3"): 30,
	}

	table, _ := New(1234, ips)

	lookupIP := netip.MustParseAddr("192.0.2.4")

	for n := 0; n < b.N; n++ {
		table.Get(lookupIP)
	}
}
