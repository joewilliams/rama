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
