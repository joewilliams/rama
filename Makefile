test:
	go test -race -v ./...

bench_rendezvous:
	go test -v -bench=. pkg/rendezvous/* -benchmem -memprofile rendezvous_memprofile.out -cpuprofile rendezvous_cpuprofile.out

bench_heavykeeper:
	go test -v -bench=. pkg/heavykeeper/* -benchmem -memprofile heavykeeper_memprofile.out -cpuprofile heavykeeper_cpuprofile.out	

bench_weighted_rendezvous:
	go test -v -bench=. pkg/weighted_rendezvous/* -benchmem -memprofile wrh_memprofile.out -cpuprofile wrh_cpuprofile.out

#go tool pprof -http localhost:3435 cpuprofile.out
#go tool pprof -http localhost:3435 memprofile.out
