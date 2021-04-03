package kvtest

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/Jonas-Heinrich/toy-distributed-key-value/kv"
)

var followerAddresses []net.IP
var leaderAddress net.IP

func TestStatus(t *testing.T) {
	// Note: This test exists on failure, since the other tests depend on it
	fmt.Println("Running test `TestStatus`..")

	var i byte
	for i = 2; true; i++ {
		ipAddress := kv.GetIPAdress(i)
		resp, err := http.Get(kv.GetURL(ipAddress, "/status"))
		if err != nil {
			fmt.Println(err)
			// Expected error when no more replicas can be found
			if i > 2 && strings.Contains(err.Error(), "connect: connection refused") {
				break
			} else {
				fmt.Println(err)
				t.Fail()
				os.Exit(1)
			}
		}
		defer resp.Body.Close()

		if !kv.TestEqualMessageResponse(resp, http.StatusOK, kv.StatusOKResponse) {
			fmt.Printf("\tStatus does not return expected answer for %s\n", ipAddress.String())
			t.Fail()
			os.Exit(1)
		} else {
			if i == 2 {
				leaderAddress = ipAddress
			} else {
				followerAddresses = append(followerAddresses, ipAddress)
			}
		}
	}
	fmt.Println("\tAll containers are up and running!")
}

func TestInitialState(t *testing.T) {
	fmt.Println("Running test `TestInitialState`..")

	if !testKVStateEqual(leaderAddress, kv.StateResponse{
		MessageResponse: kv.StatusOKResponse,
		KeyValueStore: kv.KeyValueStore{
			LocalAddress:      kv.LEADER_IP_ADDRESS,
			LeaderAddress:     kv.LEADER_IP_ADDRESS,
			FollowerAddresses: make([]net.IP, 0),
			Leader:            true,
			Initialized:       true,
			Database:          make(map[string]string),
		}}) {
		fmt.Println("\tLeader state does not match expectations")
		t.Fail()
		return
	}

	for _, elem := range followerAddresses {
		if !testKVStateEqual(elem, kv.StateResponse{
			MessageResponse: kv.StatusOKResponse,
			KeyValueStore: kv.KeyValueStore{
				LocalAddress:      elem,
				LeaderAddress:     nil,
				FollowerAddresses: make([]net.IP, 0),
				Leader:            false,
				Initialized:       false,
				Database:          make(map[string]string),
			}}) {
			fmt.Println("\tFollower state does not match expectations")
			t.Fail()
			return
		}
	}

	fmt.Println("\tAll containers have their expected initial state!")
}
