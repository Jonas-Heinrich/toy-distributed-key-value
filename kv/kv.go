package kv

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"reflect"

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

//
//
//

func (kv *KeyValueStore) Start(release bool) {
	if release {
		if leaderAddress := kv.register(kv.LeaderAddress); leaderAddress != nil {
			kv.LeaderAddress = leaderAddress
		} else {
			fmt.Println("Could not register with leader")
			os.Exit(1)
		}
	}

	r := mux.NewRouter()

	if !release {
		s := r.PathPrefix("/dev").Subrouter()
		s.HandleFunc("/kill", handleDevKill).Methods("POST")
		s.HandleFunc("/state", kv.handleDevState).Methods("GET")
		s.HandleFunc("/register", kv.handleDevRegister).Methods("POST")
	}

	r.HandleFunc("/status", handleStatus).Methods("GET")
	r.HandleFunc("/register", kv.handleRegister).Methods("POST")

	fmt.Println("Start serving..")
	http.ListenAndServe(":8080", r)
}

//
// Network Administration
//

func (kv *KeyValueStore) register(entryAddress net.IP) net.IP {
	retries := 0
	for {
		form := url.Values{}
		form.Add("ip", kv.LocalAddress.String())
		resp, err := http.PostForm(GetURL(entryAddress, "/register"), form)
		if err != nil {
			fmt.Println(err)
			return nil
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			if TestEqualMessageResponse(resp, http.StatusOK, StatusOKResponse) {
				return entryAddress
			} else {
				return nil
			}
		} else if resp.StatusCode == http.StatusServiceUnavailable {
			var response IPResponse
			bodyBytes, _ := ioutil.ReadAll(resp.Body)
			err := json.Unmarshal(bodyBytes, &response)
			if err != nil {
				return nil
			}

			if reflect.DeepEqual(response.MessageResponse, StatusMovedResponse) {
				entryAddress = response.IP
			} else {
				return nil
			}
		} else {
			return nil
		}

		retries++
		if retries > MAX_RETRIES {
			return nil
		}
	}
}
