package heavykeeper

import (
	"fmt"
	"math/rand"
	"net/netip"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

/*
	because of the width*depth memory constraint and that
	this is a probabalistic data structure we can't be 100% right all the time
	errorBound is used in the large tests to make sure we are always within
	a small percentage of actual counts
*/

const (
	errorBound = 0.9
)

func TestSmallIP(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

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

	topk := New(5, 10, 5, 0.9)

	for _, ip := range ips {
		topk.AddIP(ip)
	}

	want := map[string]uint64{
		"192.0.2.6":     8,
		"192.0.2.3":     5,
		"192.0.2.4":     4,
		"192.0.2.7":     3,
		"192.0.2.5":     2,
		"192.0.100.100": 0,
	}

	get := topk.GetIPs()
	//fmt.Println(topk.Get())
	for k, v := range want {
		assert.Equal(t, v, get[netip.MustParseAddr(k)])
	}

	wantRank := []netip.Addr{
		netip.MustParseAddr("192.0.2.6"),
		netip.MustParseAddr("192.0.2.3"),
		netip.MustParseAddr("192.0.2.4"),
		netip.MustParseAddr("192.0.2.7"),
		netip.MustParseAddr("192.0.2.5"),
	}

	//fmt.Println(topk.Rank())
	assert.Equal(t, wantRank, topk.RankIPs())
}

func TestSmallBytes(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

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

	topk := New(5, 10, 5, 0.9)

	for _, ip := range ips {
		topk.AddBytes(ip.AsSlice())
	}

	wantB := [][]byte{
		netip.MustParseAddr("192.0.2.6").AsSlice(),
		netip.MustParseAddr("192.0.2.3").AsSlice(),
		netip.MustParseAddr("192.0.2.4").AsSlice(),
		netip.MustParseAddr("192.0.2.7").AsSlice(),
		netip.MustParseAddr("192.0.2.5").AsSlice(),
	}

	wantC := []uint64{8, 5, 4, 3, 2}

	b, c := topk.RankBytes()
	//fmt.Println(b)
	//fmt.Println(c)
	assert.Equal(t, wantB, b)
	assert.Equal(t, wantC, c)
}

func TestLargeIPs(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	// this test runs multiple times to attempt to trigger nondeterministic bugs
	for i := 0; i < 500; i++ {
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

		// 20 width, 10 depth is pretty small but big enough to have decent error bounds
		topk := New(5, 20, 10, 0.9)

		for k, v := range testMap {
			for i := 0; i < v; i++ {
				topk.AddIP(netip.MustParseAddr(k))
			}
		}

		for k, v := range topk.GetIPs() {
			count := testMap[k.String()]
			p := float64(v) / float64(count)
			//fmt.Println(p)
			assert.GreaterOrEqual(t, p, errorBound)
		}
	}
}

func TestLargeBytes(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	// this test runs multiple times to attempt to trigger nondeterministic bugs
	for i := 0; i < 500; i++ {
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

		// 20 width, 10 depth is pretty small but big enough to have decent error bounds
		topk := New(5, 20, 10, 0.9)

		for k, v := range testMap {
			for i := 0; i < v; i++ {
				topk.AddBytes([]byte(k))
			}
		}

		b, c := topk.RankBytes()
		//fmt.Println(b)
		//fmt.Println(c)
		for i := 0; i < len(b); i++ {
			count := testMap[string(b[i])]
			p := float64(c[i]) / float64(count)
			//fmt.Println(p)
			assert.GreaterOrEqual(t, p, errorBound)
		}
	}
}

func BenchmarkAddIP(b *testing.B) {
	rand.Seed(time.Now().UnixNano())

	topk := New(5, 100, 100, 0.99)
	ip := netip.MustParseAddr(fmt.Sprintf("192.0.2.%v", rand.Intn(255-0)+0))

	for n := 0; n < b.N; n++ {
		topk.AddIP(ip)
	}
}

func BenchmarkRank(b *testing.B) {
	rand.Seed(time.Now().UnixNano())

	topk := New(100, 100, 100, 0.99)

	for n := 0; n < b.N; n++ {
		ip := netip.MustParseAddr(fmt.Sprintf("192.0.2.%v", rand.Intn(255-0)+0))
		topk.AddIP(ip)
		topk.RankIPs()
	}
}

func BenchmarkAddBytes(b *testing.B) {
	rand.Seed(time.Now().UnixNano())

	topk := New(5, 100, 100, 0.99)
	ip := netip.MustParseAddr(fmt.Sprintf("192.0.2.%v", rand.Intn(255-0)+0))

	for n := 0; n < b.N; n++ {
		topk.AddBytes(ip.AsSlice())
	}
}

func BenchmarkBytes(b *testing.B) {
	rand.Seed(time.Now().UnixNano())

	topk := New(100, 100, 100, 0.99)

	for n := 0; n < b.N; n++ {
		ip := netip.MustParseAddr(fmt.Sprintf("192.0.2.%v", rand.Intn(255-0)+0))
		topk.AddBytes(ip.AsSlice())
		topk.RankBytes()
	}
}
