package kv

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"reflect"
)

// Get preferred outbound ip of this machine
func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "leader:8080")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

// Get the IP address of the node with the last byte s
func GetIPAdress(s byte) net.IP {
	ipAddress := make(net.IP, len(BASE_IP_ADDRESS))
	copy(ipAddress, BASE_IP_ADDRESS)
	ipAddress[len(ipAddress)-1] = s
	return ipAddress
}

// GetURL returns the url based on an address and the path
func GetURL(ip net.IP, path string) string {
	return "http://" + ip.String() + PORT + path
}

func RespondJSON(w http.ResponseWriter, statusCode int, response interface{}) {
	payload, err := json.Marshal(response)
	if err != nil {
		log.Println(err)
		return
	}

	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")
	w.Write(payload)
}

func TestEqualMessageResponse(resp *http.Response, expectedStatusCode int, expectedResponse InfoMessage) bool {
	if resp.StatusCode != expectedStatusCode {
		fmt.Printf("Status code does not match (%d)\n", resp.StatusCode)
		return false
	}

	var actualResponse InfoMessage
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	err := json.Unmarshal(bodyBytes, &actualResponse)
	if err != nil {
		fmt.Printf("Could not parse response body\n")
		return false
	}

	equal := reflect.DeepEqual(actualResponse, expectedResponse)

	if !equal {
		actualJSON, _ := json.Marshal(actualResponse)
		expectedJSON, _ := json.Marshal(expectedResponse)
		fmt.Println("\tActual Response: " + string(actualJSON))
		fmt.Println("\tExpected Response: " + string(expectedJSON))
	}

	return equal
}

func TestEqualStateResponse(resp *http.Response, expectedStatusCode int, expectedResponse StateMessage) bool {
	if resp.StatusCode != expectedStatusCode {
		fmt.Printf("Status code does not match (%d)\n", resp.StatusCode)
		return false
	}

	var actualResponse StateMessage
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	err := json.Unmarshal(bodyBytes, &actualResponse)
	if err != nil {
		fmt.Printf("Could not parse response body\n")
		return false
	}

	equal := reflect.DeepEqual(actualResponse, expectedResponse)

	if !equal {
		actualJSON, _ := json.Marshal(actualResponse)
		expectedJSON, _ := json.Marshal(expectedResponse)
		fmt.Println("\tActual Response: " + string(actualJSON))
		fmt.Println("\tExpected Response: " + string(expectedJSON))
	}

	return equal
}
