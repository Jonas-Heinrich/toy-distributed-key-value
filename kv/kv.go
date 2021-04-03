package kv

import (
	"fmt"
	"net"
	"net/http"

	"github.com/gorilla/mux"
)

type KeyValueStore struct {
	LocalAddress      net.IP            `json:"localAddress"`
	LeaderAddress     net.IP            `json:"leaderAddress"`
	FollowerAddresses []net.IP          `json:"followerAddresses"`
	Leader            bool              `json:"leader"`
	Initialized       bool              `json:"initialized"`
	Database          map[string]string `json:"database"`
}

func InitKeyValueStore(leader bool, leaderAddress net.IP) KeyValueStore {
	localAddress := GetOutboundIP()
	if leader {
		leaderAddress = localAddress
	}
	return KeyValueStore{
		localAddress,
		leaderAddress,
		make([]net.IP, 0),
		leader,
		leader,
		make(map[string]string)}
}

func (kv KeyValueStore) Serve(release bool) {
	r := mux.NewRouter()

	if !release {
		s := r.PathPrefix("/dev").Subrouter()
		s.HandleFunc("/kill", kill).Methods("POST")
		s.HandleFunc("/state", kv.state).Methods("GET")
	}

	r.HandleFunc("/status", status).Methods("GET")

	fmt.Println("Start serving..")
	http.ListenAndServe(":8080", r)
}
