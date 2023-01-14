test:
	go test -race -v ./...

bench:
	go test -bench=. pkg/rendezvous/* -benchmem -memprofile rendezvous_memprofile.out -cpuprofile rendezvous_cpuprofile.out
	go test -bench=. pkg/heavykeeper/* -benchmem -memprofile heavykeeper_memprofile.out -cpuprofile heavykeeper_cpuprofile.out	
	#go tool pprof -http localhost:3435 cpuprofile.out
	#go tool pprof -http localhost:3435 memprofile.out