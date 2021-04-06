package kvtest

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"testing"
	"time"

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

	if !kv.TestEqualMessageResponse(resp, 200, kv.StatusOKMessage) {
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

	time.Sleep(kv.MAX_ELECTION_TIMEOUT)

	// Test Leader State
	if !testKVStateEqual(leaderAddress, kv.StateMessage{
		InfoMessage: kv.StatusOKMessage,
		KeyValueStore: kv.KeyValueStore{
			Leader: true,
			Term:   0,

			LocalAddress:      kv.LEADER_IP_ADDRESS,
			LeaderAddress:     kv.LEADER_IP_ADDRESS,
			FollowerAddresses: []net.IP{externalAddress},

			Initialized: true,
			Database:    map[string]string{"initial": "value"},
		}}) {
		fmt.Println("\tLeader state does not match expectations")
		t.Fail()
		return
	}

	// Test Follower State
	if !testKVStateEqual(externalAddress, kv.StateMessage{
		InfoMessage: kv.StatusOKMessage,
		KeyValueStore: kv.KeyValueStore{
			Leader: false,
			Term:   0,

			LocalAddress:      externalAddress,
			LeaderAddress:     leaderAddress,
			FollowerAddresses: []net.IP{externalAddress},

			Initialized: false,
			Database:    map[string]string{"initial": "value"},
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

	time.Sleep(kv.MAX_ELECTION_TIMEOUT)

	// Test Leader State
	if !testKVStateEqual(leaderAddress, kv.StateMessage{
		InfoMessage: kv.StatusOKMessage,
		KeyValueStore: kv.KeyValueStore{
			Leader: true,
			Term:   0,

			LocalAddress:      kv.LEADER_IP_ADDRESS,
			LeaderAddress:     kv.LEADER_IP_ADDRESS,
			FollowerAddresses: []net.IP{entryAddress, externalAddress},

			Initialized: true,
			Database:    map[string]string{"initial": "value"},
		}}) {
		fmt.Println("\tLeader state does not match expectations")
		t.Fail()
		return
	}

	// Test State of current followers
	for _, followerAddress := range []net.IP{entryAddress, externalAddress} {
		if !testKVStateEqual(followerAddress, kv.StateMessage{
			InfoMessage: kv.StatusOKMessage,
			KeyValueStore: kv.KeyValueStore{
				Leader: false,
				Term:   0,

				LocalAddress:      followerAddress,
				LeaderAddress:     leaderAddress,
				FollowerAddresses: []net.IP{entryAddress, externalAddress},

				Initialized: false,
				Database:    map[string]string{"initial": "value"},
			}}) {
			fmt.Println("HERE")
			fmt.Println("\tFollower state does not match expectations")
			t.Fail()
			return
		}
	}

	// Register remaining followers directly
	for _, followerAddress := range followerAddresses[2:] {
		if !testNetworkEntry(followerAddress, leaderAddress) {
			fmt.Println("\tRegistration failed")
			t.Fail()
			return
		}

	}
	fmt.Println("\tNode entered successfully!")
}
