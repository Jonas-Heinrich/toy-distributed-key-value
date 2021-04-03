package kvtest

import (
	"fmt"
	"net"
	"net/http"

	"github.com/Jonas-Heinrich/toy-distributed-key-value/kv"
)

func testKVStateEqual(address net.IP, expectedState kv.StateMessage) bool {
	resp, err := http.Get(kv.GetURL(address, "/dev/state"))
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer resp.Body.Close()

	return kv.TestEqualStateResponse(resp, http.StatusOK, expectedState)
}
