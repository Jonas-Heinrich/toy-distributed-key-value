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

var term uint64 = 0
var database = map[string]string{"initial": "value"}
var databaseLog []*kv.KeyValueLog = []*kv.KeyValueLog{kv.INITIAL_LOG}
var followers []kv.Follower
var leaderAddress net.IP

func TestStatus(t *testing.T) {
	// Note: This test exists on failure, since the other tests depend on it
	fmt.Println("Running test `TestStatus`..")

	var i byte
	for i = 2; true; i++ {
		ipAddress := kv.GetIPAdress(i)
		resp, err := http.Get(kv.GetURL(ipAddress, "/status"))
		if err != nil {
			kv.ErrorLogger.Println(err)
			// Expected error when no more replicas can be found
			if i > 2 && strings.Contains(err.Error(), "connect: connection refused") {
				break
			} else {
				kv.ErrorLogger.Println(err)
				t.Fail()
				os.Exit(1)
			}
		}
		defer resp.Body.Close()

		if !kv.TestEqualMessageResponse(resp, http.StatusOK, kv.StatusOKMessage) {
			fmt.Printf("\tStatus does not return expected answer for %s\n", ipAddress.String())
			t.Fail()
			os.Exit(1)
		} else {
			if i == 2 {
				leaderAddress = ipAddress
			} else {
				followers = append(followers, kv.Follower{
					Address:             ipAddress,
					LastLogHash:         kv.INITIAL_LOG.Hash,
					LastCommitedLogHash: kv.INITIAL_LOG.Hash,
				})
			}
		}
	}
	fmt.Println("\tAll containers are up and running!")
}

func TestInitialState(t *testing.T) {
	fmt.Println("Running test `TestInitialState`..")

	if !testLeaderState(make([]kv.Follower, 0)) {
		kv.ErrorLogger.Println("\tLeader state does not match expectations")
		t.Fail()
		return
	}

	for _, follower := range followers {
		if !testKVStateEqual(follower.Address,
			kv.StateMessage{
				InfoMessage: kv.StatusOKMessage,
				KeyValueStore: kv.KeyValueStore{
					Term:          term,
					Leader:        false,
					LeaderAddress: nil,
					Followers:     make([]kv.Follower, 0),
					LocalAddress:  follower.Address,

					Initialized: false,
					Database:    database,
					DatabaseLog: databaseLog,
				}}) {
			kv.ErrorLogger.Println("\tFollower states do not match expectations")
			t.Fail()
			return
		}
	}

	fmt.Println("\tAll containers have their expected initial state!")

	fmt.Printf(`
	Test Configuration Information
	==============================

	Leader IP Address: %s
	Follower States:   %s
`,
		leaderAddress.String(),
		followers)
}
