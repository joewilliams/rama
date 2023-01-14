package heavykeeper

import (
	"fmt"
	"math/rand"
	"net/netip"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewAndAdd(t *testing.T) {
	ips := []netip.Addr{
		netip.MustParseAddr("192.168.1.1"),
		netip.MustParseAddr("192.168.1.2"),
		netip.MustParseAddr("192.168.1.3"),
		netip.MustParseAddr("192.168.1.3"),
		netip.MustParseAddr("192.168.1.3"),
		netip.MustParseAddr("192.168.1.3"),
		netip.MustParseAddr("192.168.1.3"),
		netip.MustParseAddr("192.168.1.4"),
		netip.MustParseAddr("192.168.1.4"),
		netip.MustParseAddr("192.168.1.4"),
		netip.MustParseAddr("192.168.1.4"),
		netip.MustParseAddr("192.168.1.5"),
		netip.MustParseAddr("192.168.1.5"),
		netip.MustParseAddr("192.168.1.6"),
		netip.MustParseAddr("192.168.1.6"),
		netip.MustParseAddr("192.168.1.6"),
		netip.MustParseAddr("192.168.1.6"),
		netip.MustParseAddr("192.168.1.6"),
		netip.MustParseAddr("192.168.1.6"),
		netip.MustParseAddr("192.168.1.6"),
		netip.MustParseAddr("192.168.1.7"),
		netip.MustParseAddr("192.168.1.7"),
		netip.MustParseAddr("192.168.1.7"),
	}

	topk := New(5, 10, 10, 0.99, 1234567812345678)

	for _, ip := range ips {
		topk.Add(ip)
	}

	get := topk.Get()

	assert.Equal(t, uint64(7), get[netip.MustParseAddr("192.168.1.6")])
	assert.Equal(t, uint64(5), get[netip.MustParseAddr("192.168.1.3")])
	assert.Equal(t, uint64(4), get[netip.MustParseAddr("192.168.1.4")])
	assert.Equal(t, uint64(3), get[netip.MustParseAddr("192.168.1.7")])
	assert.Equal(t, uint64(2), get[netip.MustParseAddr("192.168.1.5")])

	want := []netip.Addr{
		netip.MustParseAddr("192.168.1.6"),
		netip.MustParseAddr("192.168.1.3"),
		netip.MustParseAddr("192.168.1.4"),
		netip.MustParseAddr("192.168.1.7"),
		netip.MustParseAddr("192.168.1.5"),
	}

	assert.Equal(t, want, topk.Rank())
}

func BenchmarkAdd(b *testing.B) {
	rand.Seed(time.Now().UnixNano())

	topk := New(100, 100, 100, 0.99, 1234567812345678)

	for n := 0; n < b.N; n++ {
		ip := func() string {
			return fmt.Sprintf("192.168.%v.%v", rand.Intn(10-0)+0, rand.Intn(255-0)+0)
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
		ip := netip.MustParseAddr(fmt.Sprintf("192.168.1.%v", rand.Intn(255-0)+0))
		topk.Add(ip)
		topk.Get()
	}
}
