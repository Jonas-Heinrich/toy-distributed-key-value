package kvtest

import (
	"net"
	"net/http"

	"github.com/Jonas-Heinrich/toy-distributed-key-value/kv"
)

func testKVStateEqual(address net.IP, expectedState kv.StateMessage) bool {
	resp, err := http.Get(kv.GetURL(address, "/dev/state"))
	if err != nil {
		kv.ErrorLogger.Println(err)
		return false
	}
	defer resp.Body.Close()

	return kv.TestEqualStateResponse(resp, http.StatusOK, expectedState)
}

//
// Test State
//

func testLeaderState(followers []kv.Follower) bool {
	return testKVStateEqual(leaderAddress,
		kv.StateMessage{
			InfoMessage: kv.StatusOKMessage,
			KeyValueStore: kv.KeyValueStore{
				Term:          term,
				Leader:        true,
				LeaderAddress: leaderAddress,
				Followers:     followers,
				LocalAddress:  leaderAddress,

				Initialized: true,
				Database:    database,
				DatabaseLog: databaseLog,
			},
		},
	)
}

func testFollowerStates(followers []kv.Follower) bool {
	for _, follower := range followers {
		if !testKVStateEqual(follower.Address,
			kv.StateMessage{
				InfoMessage: kv.StatusOKMessage,
				KeyValueStore: kv.KeyValueStore{
					Term:          term,
					Leader:        false,
					LeaderAddress: leaderAddress,
					Followers:     followers,
					LocalAddress:  follower.Address,

					Initialized: false,
					Database:    database,
					DatabaseLog: databaseLog,
				}}) {
			kv.ErrorLogger.Println("\tFollower states do not match expectations")
			return false
		}
	}
	return true
}
