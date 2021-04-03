module github.com/Jonas-Heinrich/toy-distributed-key-value

go 1.16

replace github.com/Jonas-Heinrich/toy-distributed-key-value/kv => ./kv

replace github.com/Jonas-Heinrich/toy-distributed-key-value/kvtest => ./test

require (
	github.com/gorilla/mux v1.8.0
	github.com/spf13/cobra v1.1.3
)
