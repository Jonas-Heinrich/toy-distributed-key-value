package main

import (
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/Jonas-Heinrich/toy-distributed-key-value/kv"
	kvtest "github.com/Jonas-Heinrich/toy-distributed-key-value/test"
)

func runTests() {
	flag.Set("test.v", "true")
	testing.Main(func(pat, str string) (bool, error) { return true, nil },
		[]testing.InternalTest{
			{"TestTesting", kvtest.TestTesting},
			{"TestStatus", kvtest.TestStatus}},
		[]testing.InternalBenchmark{},
		[]testing.InternalExample{})
}

func main() {
	arg := ""
	if len(os.Args) > 1 {
		arg = os.Args[1]
	} else {
		fmt.Println("Number of arguments too low.")
		os.Exit(1)
	}

	leader := false
	if arg == "leader" {
		leader = true
	} else if arg == "follower" {
		leader = false
	} else if arg == "tester" {
		runTests()
		os.Exit(0)
	} else {
		fmt.Println("Unknown command line argument.")
		os.Exit(1)
	}

	keyValueStore := kv.InitKeyValueStore(leader)
	keyValueStore.PrintKeyValueStoreInfo()
	keyValueStore.ServeKeyValueStore()
}
