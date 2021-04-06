package kvtest

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Jonas-Heinrich/toy-distributed-key-value/kv"
)

func TestLeaderElection(t *testing.T) {
	fmt.Println("Running test `TestLeaderElection`..")

	// Wait for any info passed in heartbeat to propagate
	time.Sleep(kv.MAX_ELECTION_TIMEOUT)

	// Kill current leader node
	resp, err := http.Post(kv.GetURL(leaderAddress, "/dev/kill"), "application/json", nil)
	if err != nil {
		// Expected error since the container exits right away
		if !strings.Contains(err.Error(), "EOF") {
			fmt.Println(err)
			t.Fail()
			return
		}
	} else {
		defer resp.Body.Close()
	}

	// Wait for poll to finish
	time.Sleep(kv.MAX_ELECTION_TIMEOUT)
	// Wait for first heartbeat to finish (followers correct)
	time.Sleep(kv.LEADER_HEART_BEAT_TIMEOUT * 2)

	// Get new Leader
	resp, err = http.Get(kv.GetURL(followerAddresses[0], "/leader"))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()
	var ipMessage kv.IPMessage
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	if err := json.Unmarshal(bodyBytes, &ipMessage); err != nil {
		fmt.Printf("Could not parse response body\n")
		return
	}
	if resp.StatusCode != http.StatusOK || ipMessage.InfoMessage != kv.StatusOKMessage {
		fmt.Printf("Leader request does not return expected result")
		return
	}

	leaderAddress = ipMessage.IP
	var oldFollower int
	for index, element := range followerAddresses {
		if element.Equal(leaderAddress) {
			oldFollower = index
		}
	}
	// Unordered remove of old follower
	followerAddresses[oldFollower] = followerAddresses[len(followerAddresses)-1]
	followerAddresses[len(followerAddresses)-1] = nil
	followerAddresses = followerAddresses[:len(followerAddresses)-1]

	// Check new Leader
	if !testKVStateEqual(leaderAddress, kv.StateMessage{
		InfoMessage: kv.StatusOKMessage,
		KeyValueStore: kv.KeyValueStore{
			Leader: true,
			Term:   2,

			LocalAddress:      leaderAddress,
			LeaderAddress:     leaderAddress,
			FollowerAddresses: followerAddresses,

			Initialized: true,
			Database:    map[string]string{"initial": "value"},
		}}) {
		fmt.Println("\tLeader state does not match expectations")
		t.Fail()
		return
	}

	// Check remaining Followers
	for _, followerAddress := range followerAddresses { // Check new Leader
		if !testKVStateEqual(followerAddress, kv.StateMessage{
			InfoMessage: kv.StatusOKMessage,
			KeyValueStore: kv.KeyValueStore{
				Leader: false,
				Term:   2,

				LocalAddress:      followerAddress,
				LeaderAddress:     leaderAddress,
				FollowerAddresses: followerAddresses,

				Initialized: false,
				Database:    map[string]string{"initial": "value"},
			}}) {
			fmt.Println("\tFollower state does not match expectations")
			t.Fail()
			return
		}
	}

	fmt.Println("\tNode elected successfully!")
}
