package kvtest

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/Jonas-Heinrich/toy-distributed-key-value/kv"
)

var addresses []net.IP
var leaderAddress net.IP

func testEqual(resp *http.Response, expectedStatusCode int, expectedResponse map[string]string) bool {
	if resp.StatusCode != expectedStatusCode {
		fmt.Printf("Status code does not match (%d)\n", resp.StatusCode)
		return false
	}

	var response = make(map[string]string)
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	err := json.Unmarshal(bodyBytes, &response)
	if err != nil {
		fmt.Printf("Could not parse response body\n")
		return false
	}

	return reflect.DeepEqual(response, expectedResponse)
}

func TestTesting(t *testing.T) {
	fmt.Println("Running test `TestTesting`..")
	if false {
		t.Fail()
	} else {
		fmt.Println("`TestTesting`: Testing setup works!")
	}
}

func TestStatus(t *testing.T) {
	// Note: This test exists on failure, since the other tests depend on it

	fmt.Println("Running test `TestStatus`..")
	expectedResponse := map[string]string{"status": "ok"}
	var i byte
	for i = 2; true; i++ {
		ipAddress := kv.GetIPAdress(i)
		resp, err := http.Get(kv.GetURL(ipAddress, "/status"))
		if err != nil {
			// Expected error when no more replicas can be found
			if strings.Contains(err.Error(), "connect: connection refused") {
				break
			} else {
				fmt.Println(err)
				t.Fail()
				os.Exit(1)
			}
		}
		defer resp.Body.Close()
		if !testEqual(resp, http.StatusOK, expectedResponse) {
			t.Fail()
			os.Exit(1)
		} else {
			if i == 2 {
				leaderAddress = ipAddress
			} else {
				addresses = append(addresses, ipAddress)
			}
		}
	}
	fmt.Println("`TestStatus`: All containers are up and running!")
}
