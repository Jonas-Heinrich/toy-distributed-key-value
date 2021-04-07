package kvtest

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/Jonas-Heinrich/toy-distributed-key-value/kv"
)

func testWrite(address net.IP, key string, value string) bool {
	return testWriteFollowers(followers, address, key, value)
}

func testWriteFollowers(followers []kv.Follower, address net.IP, key string, value string) bool {
	resp, err := http.Post(kv.GetURL(address, "/write/"+key), "text", bytes.NewBuffer([]byte(value)))
	if err != nil {
		fmt.Println("\tWrite request failed")
		return false
	}
	defer resp.Body.Close()

	if !kv.TestEqualMessageResponse(resp, http.StatusOK, kv.StatusOKMessage) {
		fmt.Println(resp.StatusCode)
		fmt.Println("\tWrite failed")
		return false
	}

	databaseLog = append(databaseLog, kv.CreateKeyValueLog(key, value, true, true))
	database[key] = value

	if !testLeaderState(followers) {
		kv.ErrorLogger.Println("\tLeader state does not match expectations")
		return false
	}

	// Wait for changes to fully propagate to every follower
	time.Sleep(kv.LEADER_HEART_BEAT_TIMEOUT)

	// Check follower state
	if !testFollowerStates(followers) {
		kv.ErrorLogger.Println("\tFollower states do not match expectations")
		return false
	}

	return true
}

func TestDirectWrite(t *testing.T) {
	fmt.Println("Running test `TestDirectWrite`..")

	if !testWrite(leaderAddress, "k1", "v1") {
		fmt.Println("\tWrite request failed")
		t.Fail()
		return
	}

	fmt.Println("\tWrite completed successfully!")
}

func TestIndirectWrite(t *testing.T) {
	fmt.Println("Running test `TestIndirectWrite`..")

	if !testWrite(leaderAddress, "k2", "v2") {
		fmt.Println("\tWrite request failed")
		t.Fail()
		return
	}

	fmt.Println("\tWrite completed successfully!")
}
