package kvtest

import (
	"fmt"
	"testing"
)

func TestTesting(t *testing.T) {
	fmt.Println("Running test `TestTesting`..")
	if false {
		t.Fail()
	} else {
		fmt.Println("\tTesting setup works!")
	}
}
