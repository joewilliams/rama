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

// TODO: nondeterministic test with:
// * a large number of entries
// * relatively small width*depth
// * make sure we are within error bounds

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
