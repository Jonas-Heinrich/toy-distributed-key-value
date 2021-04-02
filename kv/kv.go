package kv

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
)

type KeyValueStore struct {
	localAddress     net.IP
	leaderAddress    net.IP
	followerAdresses []net.IP
	leader           bool
	initialized      bool
	database         map[string]string
}

// Get preferred outbound ip of this machine
// Source: https://stackoverflow.com/questions/23558425/how-do-i-get-the-local-ip-address-in-go
func getOutboundIP() net.IP {
	conn, err := net.Dial("udp", "leader:8080")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

func (kv KeyValueStore) PrintKeyValueStoreInfo() {
	msg := "KeyValueStore Information\n" +
		"=========================\n" +
		"Local IP Address: %s\n" +
		"Leader: %t\n" +
		"Initialized: %t\n" +
		"Content: %s\n"

	content, err := json.MarshalIndent(kv.database, "", "    ")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Printf(msg+"\n", kv.localAddress, kv.leader, kv.initialized, content)
}

func InitKeyValueStore(leader bool) KeyValueStore {
	ipAddress := getOutboundIP()
	return KeyValueStore{
		ipAddress,
		getOutboundIP(),
		make([]net.IP, 0),
		leader,
		leader,
		make(map[string]string)}
}

func (kv KeyValueStore) ServeKeyValueStore() {
	http.HandleFunc("/status", status)
	http.HandleFunc("/kill", kill)
	http.HandleFunc("/content", kv.content)

	fmt.Println("Start serving..")
	http.ListenAndServe(":8080", nil)
}
