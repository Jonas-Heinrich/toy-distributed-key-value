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
	kv.InfoLogger.Println("Running test `TestLeaderElection`..")

	// Wait for any info passed in heartbeat to propagate
	time.Sleep(kv.LEADER_HEART_BEAT_TIMEOUT)

	// Kill current leader node
	resp, err := http.Post(kv.GetURL(leaderAddress, "/dev/kill"), "application/json", nil)
	if err != nil {
		// Expected error since the container exits right away
		if !strings.Contains(err.Error(), "EOF") {
			kv.ErrorLogger.Println(err)
			t.Fail()
			return
		}
	} else {
		defer resp.Body.Close()
	}

	// Wait for poll to finish
	time.Sleep(kv.MAX_ELECTION_TIMEOUT)
	// Wait for first heartbeat to finish (followers correct)
	time.Sleep(kv.LEADER_HEART_BEAT_TIMEOUT)

	// Get new Leader
	resp, err = http.Get(kv.GetURL(followers[0].Address, "/leader"))
	if err != nil {
		kv.ErrorLogger.Println(err)
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
	for index, follower := range followers {
		if follower.Address.Equal(leaderAddress) {
			oldFollower = index
		}
	}
	// Unordered remove of old follower
	followers[oldFollower] = followers[len(followers)-1]
	followers = followers[:len(followers)-1]

	term++

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

	kv.InfoLogger.Println("\tNode elected successfully!")
}
