package kv

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"time"

	"github.com/gorilla/mux"
)

type KeyValueStore struct {
	Leader              bool `json:"leader"`
	Term                int  `json:"term"`
	lastLeaderHeartBeat time.Time
	nextVoteTerm        int

	LocalAddress      net.IP   `json:"localAddress"`
	LeaderAddress     net.IP   `json:"leaderAddress"`
	FollowerAddresses []net.IP `json:"followerAddresses"`

	Initialized bool              `json:"initialized"`
	Database    map[string]string `json:"database"`
}

func InitKeyValueStore(leader bool, leaderAddress net.IP) KeyValueStore {
	localAddress := GetOutboundIP()
	if leader {
		leaderAddress = localAddress
	}

	return KeyValueStore{
		Leader:              leader,
		Term:                0,
		lastLeaderHeartBeat: time.Now(),
		nextVoteTerm:        0,

		LocalAddress:      localAddress,
		LeaderAddress:     leaderAddress,
		FollowerAddresses: make([]net.IP, 0),

		Initialized: leader,
		Database:    make(map[string]string),
	}
}

//
// Route Registration
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

	go kv.heartBeat()

	r := mux.NewRouter()

	if !release {
		s := r.PathPrefix("/dev").Subrouter()
		s.HandleFunc("/kill", handleDevKill).Methods("POST")
		s.HandleFunc("/state", kv.handleDevState).Methods("GET")
		s.HandleFunc("/register", kv.handleDevRegister).Methods("POST")
	}

	r.HandleFunc("/status", handleStatus).Methods("GET")

	r.HandleFunc("/register", kv.handleRegister).Methods("POST")
	r.HandleFunc("/heart-beat", kv.handleHeartBeat).Methods("POST")

	// Election
	r.HandleFunc("/poll", kv.handlePoll).Methods("GET")
	r.HandleFunc("/leader", kv.handleLeaderUpdate).Methods("POST")
	r.HandleFunc("/leader", kv.handleLeaderRequest).Methods("GET")

	fmt.Println("Start serving..")
	http.ListenAndServe(":8080", r)
}

//
// Network Administration
//

func (kv *KeyValueStore) heartBeat() {
	if !kv.Leader {
		return
	}

	for {
		for _, followerAddress := range kv.FollowerAddresses {
			go func(followerAddress net.IP) {
				jsonValue, _ := json.Marshal(HeartBeatMessage{
					InfoMessage:       StatusOKMessage,
					Term:              kv.Term,
					FollowerAddresses: kv.FollowerAddresses,
				})
				resp, err := http.Post(GetURL(followerAddress, "/heart-beat"), "application/json", bytes.NewBuffer(jsonValue))
				if err != nil {
					fmt.Println(err)
					return
				}
				defer resp.Body.Close()
			}(followerAddress)
		}

		time.Sleep(LEADER_HEART_BEAT_TIMEOUT)
	}
}

func (kv *KeyValueStore) broadCastLeaderUpdate() {
	for _, followerAddress := range kv.FollowerAddresses {
		go func(followerAddress net.IP) {
			jsonValue, _ := json.Marshal(LeaderUpdateMessage{
				Leader: kv.LeaderAddress,
				Term:   kv.Term,
			})
			resp, err := http.Post(GetURL(followerAddress, "/leader"), "application/json", bytes.NewBuffer(jsonValue))
			if err != nil {
				fmt.Println(err)
				return
			}
			defer resp.Body.Close()
		}(followerAddress)
	}
}

func (kv *KeyValueStore) runPoll() {
	kv.Term++
	yesVotes := 0
	noVotes := 0
	localAddressIndex := 0

	fmt.Printf("Running election (%d)\n", kv.Term)

	for index, followerAddress := range kv.FollowerAddresses {
		// Do not send poll to oneself
		if net.IP.Equal(followerAddress, kv.LocalAddress) {
			localAddressIndex = index
			continue
		}
		// Send poll requests to others
		go func(followerAddress net.IP, yesVotes *int, noVotes *int) {
			jsonValue, _ := json.Marshal(PollRequestMessage{
				Term:             kv.Term,
				NewLeaderAddress: kv.LocalAddress,
			})

			req, err := http.NewRequest("GET", GetURL(followerAddress, "/poll"), nil)
			if err != nil {
				log.Print(err)
				os.Exit(1)
			}
			q := req.URL.Query()
			q.Add("poll_parameters", string(jsonValue))
			req.URL.RawQuery = q.Encode()

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				fmt.Println(err)
				return
			}
			defer resp.Body.Close()

			// Evaluate poll
			bodyBytes, _ := ioutil.ReadAll(resp.Body)
			var pollResponse PollResponseMessage
			err = json.Unmarshal(bodyBytes, &pollResponse)
			if err != nil {
				fmt.Println(err)
				return
			}

			if pollResponse.Yes {
				*yesVotes += 1
			} else {
				*noVotes += 1
			}
		}(followerAddress, &yesVotes, &noVotes)
	}

	maxVotes := len(kv.FollowerAddresses) - 1
	majorityVote := int(float32(maxVotes)*0.5) + 1 // Half plus one
	won := false
	for {
		if yesVotes >= majorityVote ||
			noVotes >= majorityVote ||
			yesVotes+noVotes == maxVotes {
			won = yesVotes > noVotes
			break
		}
		time.Sleep(100 * time.Microsecond)
	}

	// Check if current election is still the newest, else invalidate
	won = won && (kv.nextVoteTerm == kv.Term)

	if won {
		kv.Leader = true
		kv.Initialized = true
		kv.LeaderAddress = kv.LocalAddress

		// Delete oneself from follower addresses
		kv.FollowerAddresses[len(kv.FollowerAddresses)-1], kv.FollowerAddresses[localAddressIndex] = kv.FollowerAddresses[localAddressIndex], kv.FollowerAddresses[len(kv.FollowerAddresses)-1]
		kv.FollowerAddresses = kv.FollowerAddresses[:len(kv.FollowerAddresses)-1]

		go kv.broadCastLeaderUpdate()
		go kv.heartBeat()

		fmt.Printf("Won election (%s)\n", kv.LocalAddress)
	}

	// If we have not won, then either another leader will contact us or the checkLeader will trigger again
}

func (kv *KeyValueStore) checkLeader() {
	// Check if leader remains alive
	for {
		if kv.Leader {
			return
		}
		if time.Since(kv.lastLeaderHeartBeat) > INDIVIDUAL_ELECTION_TIMEOUT {
			go kv.runPoll()
		}
		time.Sleep(INDIVIDUAL_ELECTION_TIMEOUT)
	}
}

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
			if TestEqualMessageResponse(resp, http.StatusOK, StatusOKMessage) {
				break
			} else {
				return nil
			}
		} else if resp.StatusCode == http.StatusServiceUnavailable {
			var response IPMessage
			bodyBytes, _ := ioutil.ReadAll(resp.Body)
			err := json.Unmarshal(bodyBytes, &response)
			if err != nil {
				return nil
			}

			if reflect.DeepEqual(response.InfoMessage, StatusMovedMessage) {
				entryAddress = response.IP
			} else {
				return nil
			}
		} else {
			return nil
		}

		retries++
		if retries > MAX_REGISTER_RETRIES {
			return nil
		}
	}

	// Reset last leader heart beat to avoid instant election
	kv.lastLeaderHeartBeat = time.Now()
	go kv.checkLeader()

	return entryAddress
}
