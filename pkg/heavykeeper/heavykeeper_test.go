package heavykeeper

import (
	"fmt"
	"math/rand"
	"net/netip"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSmall(t *testing.T) {
	ips := []netip.Addr{
		netip.MustParseAddr("192.0.2.6"),
		netip.MustParseAddr("2001:0db8:85a3:1:1:8a2e:0370:7334"),
		netip.MustParseAddr("192.0.2.2"),
		netip.MustParseAddr("192.0.2.3"),
		netip.MustParseAddr("192.0.2.3"),
		netip.MustParseAddr("192.0.2.3"),
		netip.MustParseAddr("192.0.2.3"),
		netip.MustParseAddr("192.0.2.3"),
		netip.MustParseAddr("192.0.2.4"),
		netip.MustParseAddr("192.0.2.4"),
		netip.MustParseAddr("192.0.2.4"),
		netip.MustParseAddr("192.0.2.4"),
		netip.MustParseAddr("192.0.2.5"),
		netip.MustParseAddr("192.0.2.5"),
		netip.MustParseAddr("192.0.2.6"),
		netip.MustParseAddr("192.0.2.6"),
		netip.MustParseAddr("192.0.2.6"),
		netip.MustParseAddr("192.0.2.6"),
		netip.MustParseAddr("192.0.2.6"),
		netip.MustParseAddr("192.0.2.6"),
		netip.MustParseAddr("192.0.2.7"),
		netip.MustParseAddr("192.0.2.7"),
		netip.MustParseAddr("192.0.2.7"),
		netip.MustParseAddr("192.0.2.6"),
	}

	// this test is deterministic until we hit width*depth
	topk := New(5, 10, 5, 0.9, 1234567812345678)

	for _, ip := range ips {
		topk.Add(ip)
	}

	want := map[string]uint64{
		"192.0.2.6":                         8,
		"192.0.2.3":                         5,
		"192.0.2.4":                         4,
		"2001:0db8:85a3:1:1:8a2e:0370:7334": 1,
		"192.0.2.5":                         2,
		"192.0.100.100":                     0,
	}

	get := topk.Get()
	//fmt.Println(topk.Get())
	for k, v := range want {
		assert.Equal(t, v, get[netip.MustParseAddr(k)])
	}

	wantRank := []netip.Addr{
		netip.MustParseAddr("192.0.2.6"),
		netip.MustParseAddr("192.0.2.3"),
		netip.MustParseAddr("192.0.2.4"),
		netip.MustParseAddr("192.0.2.5"),
		netip.MustParseAddr("2001:0db8:85a3:1:1:8a2e:0370:7334"),
	}

	//fmt.Println(topk.Rank())
	assert.Equal(t, wantRank, topk.Rank())
}

func TestLarge(t *testing.T) {
	testMap := map[string]int{
		"192.0.2.1":                         1000,
		"192.0.2.2":                         5000,
		"192.0.2.3":                         100,
		"2001:0db8:85a3:1:1:8a2e:0370:7334": 300,
		"192.0.2.100":                       50,
		"192.0.2.101":                       10,
		"192.0.2.200":                       1,
		"192.0.2.201":                       75,
		"192.0.2.170":                       25,
		"192.0.2.65":                        500,
		"192.0.2.34":                        2000,
		"192.0.2.122":                       1200,
		"192.0.2.111":                       10,
		"192.0.2.12":                        80,
		"192.0.2.113":                       800,
		"192.0.2.114":                       90,
		"192.0.2.15":                        123,
		"192.0.2.116":                       234,
		"192.0.2.117":                       345,
		"192.0.2.118":                       85,
		"192.0.2.21":                        8,
	}

	rand.Seed(time.Now().UnixNano())

	topk := New(5, 100, 5, 0.9, 1234567812345678)

	for k, v := range testMap {
		for i := 0; i < v; i++ {
			topk.Add(netip.MustParseAddr(k))
		}
	}

	//fmt.Println(topk.Get())
}

func BenchmarkAdd(b *testing.B) {
	rand.Seed(time.Now().UnixNano())

	topk := New(100, 100, 100, 0.99, 1234567812345678)

	for n := 0; n < b.N; n++ {
		ip := func() string {
			return fmt.Sprintf("192.0.%v.%v", rand.Intn(10-0)+0, rand.Intn(255-0)+0)
		}
		topk.Add(netip.MustParseAddr(ip()))
		topk.Add(netip.MustParseAddr(ip()))
		topk.Add(netip.MustParseAddr(ip()))
		topk.Add(netip.MustParseAddr(ip()))
		topk.Add(netip.MustParseAddr(ip()))
		topk.Add(netip.MustParseAddr(ip()))
		topk.Add(netip.MustParseAddr(ip()))
		topk.Add(netip.MustParseAddr(ip()))
		topk.Add(netip.MustParseAddr(ip()))
		topk.Add(netip.MustParseAddr(ip()))
	}
}

func BenchmarkGet(b *testing.B) {
	rand.Seed(time.Now().UnixNano())

	topk := New(100, 100, 100, 0.99, 1234567812345678)

	for n := 0; n < b.N; n++ {
		ip := netip.MustParseAddr(fmt.Sprintf("192.0.2.%v", rand.Intn(255-0)+0))
		topk.Add(ip)
		topk.Get()
	}
}
