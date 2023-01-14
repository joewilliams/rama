## Rendezvous with Rama

This repo contains hash table implementations in golang. These implementations are usually focused on use in network related services (i.e. load balancers, etc) and storing stuff like IP addresses rather than any other type of data.

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

I've done a little bit of profiling and had the following observations:
* The radix sort from https://github.com/twotwotwo/sorts seems a bit faster than [sort.Slice](https://pkg.go.dev/sort#Slice) and [slices.Sort](https://pkg.go.dev/golang.org/x/exp/slices#Sort)
* Unsurprisingly `binary.LittleEndian.PutUint32(bI, uint32(i))` seems to be a lot faster than `[]byte(fmt.Sprint())`
* Previously this used [siphash](https://en.wikipedia.org/wiki/SipHash) but for this use case I think a seeded [xxhash](https://cyan4973.github.io/xxHash/) is equivalently safe in terms of collisions and hash flood resistance and is a bit faster. Hash speed is not a huge factor in this use case though.

There's probably more that can be done but these were the obvious things.


### HeavyKeeper

[HeavyKeeper](https://www.usenix.org/system/files/conference/atc18/atc18-gong.pdf) is a probabilistic data structure for maintaining a top-k dataset. It improves upon previous top-k implementations in speed and accuracy by using something called *count-with-exponential-decay*, which basically means entries in the dataset are heavily biased towards high frequency i.e. entries we rarely see are quicky replaced by entries we see very often. Multiple hash tables ("buckets") are used to improve accuracy by storing counts multiple times and picking the largest. The data structure is tunable in terms of the size of `k` as well as performance, memory usage and accuracy which are determined by `width`, `depth` and `decay`. Higher values for each tend to use more cpu and memory but will be more accurate. For instance higher values for `width` and `depth` will mean there is a better chance the "correct" count is stored somewhere for a given entry but results in larger hash tables and more iterations through those tables. `decay` controls how much bias there is, higher values will mean rare entries are removed more quickly. `New` creates a new instance, `Add` adds an entry to the data structure, `Get` returns a map of the current top-k entries and their counts and `Rank` returns a sorted array of the top-k entries.

This implementation was inspired by the [original C++ implementation](https://github.com/papergitkeeper/heavy-keeper-project/) and a [golang implementation](https://github.com/migotom/heavykeeper). In the implementation here I have focused on only storing the `netip.Addr` type. This allows some assumptions to be made in the underlying data structures. I have made improvements to reduce memory allocations and hashing, and finding a specific entry in the minheap.

```
hashKey := 1234567812345678
k := 5
width := 10
depth := 3
decay := 0.9

topk := New(k, width, depth, decay, hashKey)

topk.Add(netip.MustParseAddr("192.168.1.4"))

results := topk.Get()
```

Profiling and performance observations:
* `minHeap.find(ip)` uses a sort and binary search (using `netip.Addr.Less()`) rather than linear search, if `k` is large this should help. `find` happens on each `Add` I think this is as good as it gets but finding a shortcut would be nice.
* I am not a fan of needing `minHeap.find(ip)` on each `Add` so to get more out of that sort and binary search I store the IP address as bytes and it's hashed fingerprint in each node in the heap. This means if an entry is in the top-k we won't have to `ip.AsSlice()` nor `xxhash.Checksum64S` when we add it. So IPs we see often have better performance than ones we see rarely.
* Under the covers `Rank()` uses the parallel radix sort I used in the rendezvous hash. Again this probably only helps when `k` is large.