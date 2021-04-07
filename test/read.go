package kvtest

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"testing"

	"github.com/Jonas-Heinrich/toy-distributed-key-value/kv"
)

func testRead(address net.IP, key string, expectedValue_ string, expectFind bool) bool {
	resp, err := http.Get(kv.GetURL(address, "/read/"+key))
	if err != nil {
		fmt.Println("\tRead request failed")
		return false
	}
	defer resp.Body.Close()

	readMessageBytes, _ := ioutil.ReadAll(resp.Body)
	var readMessage kv.ReadMessage
	err = json.Unmarshal(readMessageBytes, &readMessage)
	if err != nil {
		fmt.Println("\tRead message format unknown")
		return false
	}

	var expectedStatusCode int
	var expectedInfoMessage kv.InfoMessage
	var expectedValue string
	if expectFind {
		expectedStatusCode = http.StatusOK
		expectedInfoMessage = kv.StatusOKMessage
		expectedValue = expectedValue_
	} else {
		expectedStatusCode = http.StatusNotFound
		expectedInfoMessage = kv.StatusValueNotFoundMessage
		expectedValue = ""
	}

	if resp.StatusCode != expectedStatusCode || readMessage.InfoMessage != expectedInfoMessage || readMessage.Value != expectedValue {
		fmt.Printf(`
		Unexpected values:

		HTTP Status Code: %d,
		InfoMessage: %s,
		Value: %s
`, resp.StatusCode, readMessage.InfoMessage, readMessage.Value)
		fmt.Println("\tRead request returned unexpected response")
		return false
	}
	return true
}

func TestInitialDirectRead(t *testing.T) {
	fmt.Println("Running test `TestInitialDirectRead`..")

	if !testRead(leaderAddress, "initial", "value", true) {
		fmt.Println("\tRead request failed")
		t.Fail()
		return
	}

	fmt.Println("\tRead completed successfully!")
}

func TestInitialIndirectRead(t *testing.T) {
	fmt.Println("Running test `TestInitialIndirectRead`..")

	if !testRead(followers[0].Address, "initial", "value", true) {
		fmt.Println("\tRead request failed")
		t.Fail()
		return
	}

	fmt.Println("\tRead completed successfully!")
}

func TestInitialDirectReadNotFound(t *testing.T) {
	fmt.Println("Running test `TestInitialDirectReadNotFound`..")

	if !testRead(leaderAddress, "whatever", "", false) {
		fmt.Println("\tRead request failed")
		t.Fail()
		return
	}

	fmt.Println("\tRead completed successfully!")
}

func TestInitialIndirectReadNotFound(t *testing.T) {
	fmt.Println("Running test `TestInitialIndirectReadNotFound`..")

	if !testRead(followers[0].Address, "whatever", "", false) {
		fmt.Println("\tRead request failed")
		t.Fail()
		return
	}

	fmt.Println("\tRead completed successfully!")
}
