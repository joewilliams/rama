## Rendezvous with Rama

This repo contains data structure implementations in golang. These implementations are usually focused on use in network related services (i.e. load balancers, etc) and storing stuff like IP addresses rather than any other type of data.

### General Approach
* __Minimalism__ - The implementations here try to do only one thing well and don't tend to be generalized. I try to use as few external dependencies as possible. When I do they tend to be for things like sorting and hashing. My goal is simple, easily understandable and straight forward implementations without too many bells or whistles.
* __Purity__ - This library is mostly procedural golang. If you need concurrency, channels, locking, etc build it on top or send me a convincing PR. Additionally, it exposes standard golang types like maps or arrays as output. I won't make you rely on the types in this library in your code.
* __Network focused__ - This library focuses on networking related use cases. Most things here use the [`netip.Addr` type](https://pkg.go.dev/net/netip#Addr), this makes the implementations simpler than being generalized.
* __Correct and Fast__ - Once I get a implementation working correctly I add tests and then make in simple and fast. If an optimization makes the code hard to understand I usually avoid it.

### Rendezvous Hash

This [rendezvous hash](https://en.wikipedia.org/wiki/Rendezvous_hashing) implementation is based on the one we did in [glb](https://github.com/github/glb-director/blob/master/docs/development/glb-hashing.md). Rather than hash and performing the ranking on each look up the table is generated with a list of IP addresses. Look ups are constant time by indexing into the table. Adding/deleting entries from the table causes a full regeneration of the table but deleting maintains the "minimal disruption" property rendezvous is commonly used for. The load balancing of the table is roughly equal between members but not exactly equal. By default the table size dynamic, based on the number of members on the first `New` call. Each member gets around 100 "slots", usually +/- 10. `NewWithTableSize` is available for picking the table size. Setting the hash key to a known value ensures the table is generated identically between discrete runs. If the key is set to zero a random value will be used which can be retrieved using `Key` for persisting across runs. It uses [xxhash](https://cyan4973.github.io/xxHash/) for hashing table members and look ups.
```
hashKey := 1234567812345678

ips := []netip.Addr{
		netip.MustParseAddr("192.168.1.1"),
		netip.MustParseAddr("192.168.1.2"),
		netip.MustParseAddr("192.168.1.3"),
	}

table, err := New(hashKey, ips)

ip := table.Get(netip.MustParseAddr("172.16.1.1"))

newEntry := netip.MustParseAddr("192.168.1.4")
table.Add(newEntry)

table.Delete(newEntry)
```

Profiling and performance observations:
* Unsurprisingly `binary.LittleEndian.PutUint32(bI, uint32(i))` seems to be a lot faster than `[]byte(fmt.Sprint())` when generating the row hash.
* Previously this used [siphash](https://en.wikipedia.org/wiki/SipHash) but for this use case I think a seeded [xxhash](https://cyan4973.github.io/xxHash/) is equivalently safe for this use case, and is a bit faster. Hash speed is not a huge factor in this use case though. 

### Weighted Rendezvous Hash

This implementation is based on the rendezvous hash described above but adds weighting to each member of the table while maintaining the "minimal disruption" property on delete. The weighting implementation is described in this [presentation](https://www.snia.org/sites/default/files/SDC15_presentations/dist_sys/Jason_Resch_New_Consistent_Hashings_Rev.pdf). It maintains the constant time look up by pre-generating the table on modification. `New` and `NewWithTableSize` now require a map of addresses and weights, as does `Add`. `Delete` and `Get` work the same. It has an additional `Set` call that allows for adjusting an existing members weight and regenerating the table.

```
ips := map[netip.Addr]float64{
		netip.MustParseAddr("192.0.2.111"): 10,
		netip.MustParseAddr("192.0.2.112"): 20,
		netip.MustParseAddr("192.0.2.113"): 70,
	}

table, err := New(1234567812345678, ips)
```

Profiling and performance observations:
* This implementation should perform nearly identically as the above. The only change is some extra math to deal with the weights. 
* Converting a uint64 to a uniformly random float64 in golang is a trick https://github.com/golang/go/issues/12290 The implementation I used is from [here](http://www.math.sci.hiroshima-u.ac.jp/m-mat/MT/emt64.html).
* Scoring each member of the table is mostly bound to how fast `math.Log` returns, this the primary reason `generateTable` is slower in the weighted version of rendezvous.
* I tried a number of things to make combining two `[]byte` together faster during table generation but didn't find anything better than `append`. Using `bytes.NewBuffer` and `bytes.Write` didn't help, nor did looping and `copy`, `bytes.Join` seemed about the same.

### HeavyKeeper

[HeavyKeeper](https://www.usenix.org/system/files/conference/atc18/atc18-gong.pdf) is a probabilistic data structure for maintaining a top-k dataset. It improves upon previous top-k implementations in speed and accuracy by using something called *count-with-exponential-decay*, which basically means entries in the dataset are heavily biased towards high frequency i.e. entries we rarely see are quicky replaced by entries we see very often. Multiple hash tables ("buckets") are used to improve accuracy by storing counts multiple times and picking the largest. The data structure is tunable in terms of the size of `k` as well as performance, memory usage and accuracy which are determined by `width`, `depth` and `decay`. Higher values for each tend to use more cpu and memory but will be more accurate. For instance higher values for `width` and `depth` will mean there is a better chance the "correct" count is stored somewhere for a given entry but results in larger hash tables and more iterations through those tables. `decay` controls how much bias there is, higher values will mean rare entries are removed more quickly. `New` and `NewWithSeed` create a new instance, `AddIP` and `AddBytes` adds an entry to the data structure, `GetIPs` returns a map of the current top-k IPs and their counts and `RankIPs` and `RankBytes` returns a sorted array(s) of the top-k entries.

This implementation was inspired by the [original C++ implementation](https://github.com/papergitkeeper/heavy-keeper-project/) and a [golang implementation](https://github.com/migotom/heavykeeper). In the implementation here I have focused on only storing the `netip.Addr` and `[]byte` types. This allows some assumptions to be made in the underlying data structures. I have made improvements to reduce memory allocations and hashing.

```
hashKey := 1234567812345678
k := 5
width := 10
depth := 3
decay := 0.9

topk := New(k, width, depth, decay, hashKey)

topk.AddIP(netip.MustParseAddr("192.168.1.4"))

results := topk.GetIPs()
```

Profiling and performance observations:
* `minHeap.find(ip)` uses a linear search, in the worst case the size of `k`. Sort and binary search might be better if `k` is large but is overkill and worse most of the time. Implementing [`slices.BinarySearchFunc`](https://pkg.go.dev/golang.org/x/exp/slices#BinarySearchFunc) using `netip.Addr.Less()` doesn't seem to work in this scenario, I think mostly due to how IPv6 and IPv4 are sorted against each other. I think linear search is as good as it gets but finding a shortcut would be nice.
* To get more out of that linear search I store the IP address as bytes and it's fingerprint in each node in the heap. This means if an entry is in the current top-k we won't have to `ip.AsSlice()` nor `xxhash.Checksum64S` when we add it. This reduces allocations quite a bit. IPs we see often have better performance than ones we see rarely.
* Under the covers `Rank()` uses [`slices.SortFunc`](https://pkg.go.dev/golang.org/x/exp/slices#SortFunc).
* `RankBytes` returns two sorted arrays rather than a map with `string([]byte)` keys.


#### To Do
* Enhancing Peak Network Traffic Prediction via Time-Series Decomposition https://arxiv.org/pdf/2303.13529.pdf
* WDHT http://archive.cone.informatik.uni-freiburg.de/pubs/WDHT.pdf (not available over HTTPS)
* SIEVE: an Efficient Turn-Key Eviction Algorithm for Web Caches https://cachemon.github.io/SIEVE-website/blog/2023/12/17/sieve-is-simpler-than-lru
* SSSP with negative weights:
  * https://anonymous.4open.science/r/negative-weight-shortest-path/README.md
  * https://arxiv.org/abs/2203.03456
  * https://arxiv.org/abs/2304.05279
  * https://arxiv.org/abs/2502.11999
