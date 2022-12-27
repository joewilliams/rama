test:
	go test -v ./...

bench:
	go test -bench=.  -benchmem -memprofile memprofile.out -cpuprofile cpuprofile.out
	
pprofcpu:
	go tool pprof -http localhost:3435 cpuprofile.out

pprofmem:
	go tool pprof -http localhost:3435 memprofile.out