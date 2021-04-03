package kvtest

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"testing"

	"github.com/Jonas-Heinrich/toy-distributed-key-value/kv"
)

func testNetworkEntry(externalAddress net.IP, entryAddress net.IP) bool {
	form := url.Values{}
	form.Add("ip", entryAddress.String())
	resp, err := http.PostForm(kv.GetURL(externalAddress, "/dev/register"), form)
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer resp.Body.Close()

	if !kv.TestEqualMessageResponse(resp, 200, kv.StatusOKResponse) {
		fmt.Println("\tNode did not return expected response")
		return false
	}
	return true
}

func TestDirectNetworkEntry(t *testing.T) {
	fmt.Println("Running test `TestDirectNetworkEntry`..")

	externalAddress := followerAddresses[0]
	entryAddress := leaderAddress

	if !testNetworkEntry(externalAddress, entryAddress) {
		fmt.Println("\tRegistration failed")
		t.Fail()
		return
	}

	// Test Leader State
	if !testKVStateEqual(leaderAddress, kv.StateResponse{
		MessageResponse: kv.StatusOKResponse,
		KeyValueStore: kv.KeyValueStore{
			LocalAddress:      kv.LEADER_IP_ADDRESS,
			LeaderAddress:     kv.LEADER_IP_ADDRESS,
			FollowerAddresses: []net.IP{externalAddress},
			Leader:            true,
			Initialized:       true,
			Database:          make(map[string]string),
		}}) {
		fmt.Println("\tLeader state does not match expectations")
		t.Fail()
		return
	}

	// Test Follower State
	if !testKVStateEqual(externalAddress, kv.StateResponse{
		MessageResponse: kv.StatusOKResponse,
		KeyValueStore: kv.KeyValueStore{
			LocalAddress:      externalAddress,
			LeaderAddress:     leaderAddress,
			FollowerAddresses: make([]net.IP, 0),
			Leader:            false,
			Initialized:       false,
			Database:          make(map[string]string),
		}}) {
		fmt.Println("\tFollower state does not match expectations")
		t.Fail()
		return
	}

	fmt.Println("\tNode entered successfully!")
}

func TestIndirectNetworkEntry(t *testing.T) {
	fmt.Println("Running test `TestIndirectNetworkEntry`..")

	externalAddress := followerAddresses[1]
	entryAddress := followerAddresses[0]

	if !testNetworkEntry(externalAddress, entryAddress) {
		fmt.Println("\tRegistration failed")
		t.Fail()
		return
	}

	// Test Leader State
	if !testKVStateEqual(leaderAddress, kv.StateResponse{
		MessageResponse: kv.StatusOKResponse,
		KeyValueStore: kv.KeyValueStore{
			LocalAddress:      kv.LEADER_IP_ADDRESS,
			LeaderAddress:     kv.LEADER_IP_ADDRESS,
			FollowerAddresses: []net.IP{entryAddress, externalAddress},
			Leader:            true,
			Initialized:       true,
			Database:          make(map[string]string),
		}}) {
		fmt.Println("\tLeader state does not match expectations")
		t.Fail()
		return
	}

	// Test Follower State
	if !testKVStateEqual(externalAddress, kv.StateResponse{
		MessageResponse: kv.StatusOKResponse,
		KeyValueStore: kv.KeyValueStore{
			LocalAddress:      externalAddress,
			LeaderAddress:     leaderAddress,
			FollowerAddresses: make([]net.IP, 0),
			Leader:            false,
			Initialized:       false,
			Database:          make(map[string]string),
		}}) {
		fmt.Println("\tFollower state does not match expectations")
		t.Fail()
		return
	}

	fmt.Println("\tNode entered successfully!")
}
