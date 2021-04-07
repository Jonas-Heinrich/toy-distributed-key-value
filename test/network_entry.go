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
		kv.ErrorLogger.Println(err)
		return false
	}
	defer resp.Body.Close()

	if !kv.TestEqualMessageResponse(resp, 200, kv.StatusOKMessage) {
		kv.ErrorLogger.Println("\tNode did not return expected response")
		return false
	}
	return true
}

func TestDirectNetworkEntry(t *testing.T) {
	fmt.Println("Running test `TestDirectNetworkEntry`..")

	externalAddress := followers[0].Address
	entryAddress := leaderAddress

	if !testNetworkEntry(externalAddress, entryAddress) {
		kv.ErrorLogger.Println("\tRegistration failed")
		t.Fail()
		return
	}

	time.Sleep(kv.MAX_ELECTION_TIMEOUT)

	// Test Leader State
	if !testKVStateEqual(leaderAddress, kv.StateMessage{
		InfoMessage: kv.StatusOKMessage,
		KeyValueStore: kv.KeyValueStore{
			Term:          term,
			Leader:        true,
			LeaderAddress: leaderAddress,
			Followers:     followers[:1],
			LocalAddress:  leaderAddress,

			Initialized: true,
			Database:    database,
			DatabaseLog: databaseLog,
		}}) {
		kv.ErrorLogger.Println("\tLeader state does not match expectations")
		t.Fail()
		return
	}

	// Test Follower State
	if !testKVStateEqual(externalAddress, kv.StateMessage{
		InfoMessage: kv.StatusOKMessage,
		KeyValueStore: kv.KeyValueStore{
			Term:          term,
			Leader:        false,
			LeaderAddress: leaderAddress,
			Followers:     followers[:1],
			LocalAddress:  externalAddress,

			Initialized: false,
			Database:    database,
			DatabaseLog: databaseLog,
		}}) {
		kv.ErrorLogger.Println("\tFollower state does not match expectations")
		t.Fail()
		return
	}

	kv.InfoLogger.Println("\tNode entered successfully!")
}

func TestIndirectNetworkEntry(t *testing.T) {
	fmt.Println("Running test `TestIndirectNetworkEntry`..")

	externalAddress := followers[1].Address
	entryAddress := followers[0].Address

	if !testNetworkEntry(externalAddress, entryAddress) {
		kv.ErrorLogger.Println("\tRegistration failed")
		t.Fail()
		return
	}

	time.Sleep(kv.MAX_ELECTION_TIMEOUT)

	// Test Leader State
	if !testKVStateEqual(leaderAddress, kv.StateMessage{
		InfoMessage: kv.StatusOKMessage,
		KeyValueStore: kv.KeyValueStore{
			Term:          term,
			Leader:        true,
			LeaderAddress: leaderAddress,
			Followers:     followers[:2],
			LocalAddress:  leaderAddress,

			Initialized: true,
			Database:    database,
			DatabaseLog: databaseLog,
		}}) {
		kv.ErrorLogger.Println("\tLeader state does not match expectations")
		t.Fail()
		return
	}

	// Test State of current followers
	for _, followerAddress := range []net.IP{entryAddress, externalAddress} {
		if !testKVStateEqual(followerAddress, kv.StateMessage{
			InfoMessage: kv.StatusOKMessage,
			KeyValueStore: kv.KeyValueStore{
				Term:          term,
				Leader:        false,
				LeaderAddress: leaderAddress,
				Followers:     followers[:2],
				LocalAddress:  followerAddress,

				Initialized: false,
				Database:    database,
				DatabaseLog: databaseLog,
			}}) {
			kv.ErrorLogger.Println("\tFollower states do not match expectations")
			t.Fail()
			return
		}
	}

	kv.InfoLogger.Println("\tNode entered successfully!")
}

func TestNetworkEntryLogReplication(t *testing.T) {
	fmt.Println("Running test `TestNetworkEntryLogReplication`..")

	testWriteFollowers(followers[:2], leaderAddress, "log", "replication")
	time.Sleep(kv.LEADER_HEART_BEAT_TIMEOUT)

	follower := followers[2]
	if !testNetworkEntry(follower.Address, leaderAddress) {
		kv.ErrorLogger.Println("\tRegistration failed")
		t.Fail()
		return
	}

	time.Sleep(kv.LEADER_HEART_BEAT_TIMEOUT)

	if !testLeaderState(followers[:3]) {
		kv.ErrorLogger.Println("\tLeader state does not match expectations")
		t.Fail()
		return
	}

	if !testFollowerStates(followers[:3]) {
		kv.ErrorLogger.Println("\tFollower states do not match expectations")
		t.Fail()
		return
	}

	kv.InfoLogger.Println("\tNode entered and logs replicated successfully!")
}

func TestRemainingNetworkEntry(t *testing.T) {
	fmt.Println("Running test `TestRemainingNetworkEntry`..")

	// Register remaining followers directly
	for _, follower := range followers[3:] {
		if !testNetworkEntry(follower.Address, leaderAddress) {
			kv.ErrorLogger.Println("\tRegistration failed")
			t.Fail()
			return
		}
	}

	time.Sleep(kv.LEADER_HEART_BEAT_TIMEOUT)

	if !testLeaderState(followers) {
		kv.ErrorLogger.Println("\tLeader state does not match expectations")
		t.Fail()
		return
	}

	if !testFollowerStates(followers) {
		kv.ErrorLogger.Println("\tFollower states do not match expectations")
		t.Fail()
		return
	}

	kv.InfoLogger.Println("\tNode entered successfully!")
}
