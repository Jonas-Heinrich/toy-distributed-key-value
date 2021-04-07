package kv

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/mux"
)

type Follower struct {
	Address net.IP `json:"address"`

	LastLogHash         string `json:"lastLogHash"`
	LastCommitedLogHash string `json:"lastCommitedLogHash"`
}

type KeyValueStore struct {
	// "Shared" Network Properties

	Term          uint64     `json:"term"`
	Leader        bool       `json:"leader"`
	LeaderAddress net.IP     `json:"leaderAddress"`
	Followers     []Follower `json:"followers"`
	LocalAddress  net.IP     `json:"localAddress"`

	// Private Network Properties

	lastLeaderHeartBeat time.Time
	nextVoteTerm        uint64

	// Database Properties

	Initialized bool              `json:"initialized"`
	Database    map[string]string `json:"database"`
	DatabaseLog []*KeyValueLog    `json:"databaseLog"`

	// Mutex

	followerMutex sync.RWMutex
	databaseMutex sync.RWMutex
	logMutex      sync.RWMutex
}

func InitKeyValueStore(leader bool, leaderAddress net.IP) KeyValueStore {
	localAddress := GetOutboundIP()
	if leader {
		leaderAddress = localAddress
	}

	return KeyValueStore{
		Term:          0,
		Leader:        leader,
		LeaderAddress: leaderAddress,
		Followers:     make([]Follower, 0),
		LocalAddress:  localAddress,

		lastLeaderHeartBeat: time.Now(),
		nextVoteTerm:        0,

		Initialized: leader,
		Database:    map[string]string{"initial": "value"},
		DatabaseLog: []*KeyValueLog{INITIAL_LOG},
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
			ErrorLogger.Println("Could not register with leader")
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

	// Read
	r.HandleFunc("/read/{key}", kv.handleRead).Methods("GET")

	// Write
	r.HandleFunc("/write/{key}", kv.handleWrite).Methods("POST")
	r.HandleFunc("/log/append", kv.handleLogAppend).Methods("POST")
	r.HandleFunc("/log/commit", kv.handleCommit).Methods("POST")

	InfoLogger.Println("Start serving..")
	http.ListenAndServe(":8080", r)
}

//
// Network Administration
//

func (kv *KeyValueStore) heartBeat() {
	if !kv.Leader {
		return
	}

	// As long as the leader lives
	for {
		// Send HeartBeat to current followers
		kv.followerMutex.RLock()
		for _, follower := range kv.Followers {
			go func(follower Follower) {
				jsonValue, _ := json.Marshal(HeartBeatMessage{
					InfoMessage: StatusOKMessage,
					Term:        kv.Term,
					Followers:   kv.Followers,
				})
				resp, err := http.Post(GetURL(follower.Address, "/heart-beat"), "application/json", bytes.NewBuffer(jsonValue))
				if err != nil {
					ErrorLogger.Println(err)
					return
				}
				defer resp.Body.Close()
			}(follower)
		}
		kv.followerMutex.RUnlock()
		time.Sleep(LEADER_HEART_BEAT_TIMEOUT)
	}
}

func (kv *KeyValueStore) runPoll() {
	atomic.AddUint64(&kv.Term, 1)
	yesVotes := 0
	noVotes := 0
	localAddressIndex := 0

	InfoLogger.Printf("Running election (%d)\n", kv.Term)

	kv.followerMutex.RLock()
	kv.logMutex.RLock()
	for index, follower := range kv.Followers {
		// Do not send poll to oneself
		if net.IP.Equal(follower.Address, kv.LocalAddress) {
			localAddressIndex = index
			continue
		}
		// Send poll requests to others
		go func(term uint64, lastLogHash string, follower Follower, yesVotes *int, noVotes *int) {
			jsonValue, _ := json.Marshal(PollRequestMessage{
				Term:             term,
				NewLeaderAddress: kv.LocalAddress,
				LastLogHash:      lastLogHash,
			})

			req, err := http.NewRequest("GET", GetURL(follower.Address, "/poll"), nil)
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
				kv.logMutex.RUnlock()
				kv.followerMutex.RUnlock()
				ErrorLogger.Println(err)
				return
			}
			defer resp.Body.Close()

			// Evaluate poll
			bodyBytes, _ := ioutil.ReadAll(resp.Body)
			var pollResponse PollResponseMessage
			err = json.Unmarshal(bodyBytes, &pollResponse)
			if err != nil {
				kv.logMutex.RUnlock()
				kv.followerMutex.RUnlock()
				ErrorLogger.Println(err)
				return
			}

			if pollResponse.Yes {
				*yesVotes += 1
			} else {
				*noVotes += 1
			}
		}(kv.Term, kv.DatabaseLog[kv.findLastCommitedLog()].Hash, follower, &yesVotes, &noVotes)
	}
	kv.logMutex.RUnlock()
	maxVotes := len(kv.Followers) - 1
	kv.followerMutex.RUnlock()

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
	won = won && (kv.nextVoteTerm <= kv.Term)

	if won {
		kv.Leader = true
		kv.Initialized = true
		kv.LeaderAddress = kv.LocalAddress

		// Delete oneself from follower addresses
		kv.followerMutex.Lock()
		kv.Followers[len(kv.Followers)-1], kv.Followers[localAddressIndex] = kv.Followers[localAddressIndex], kv.Followers[len(kv.Followers)-1]
		kv.Followers = kv.Followers[:len(kv.Followers)-1]
		kv.followerMutex.Unlock()

		// Broadcast leader update
		leaderData := LeaderUpdateMessage{
			Leader: kv.LeaderAddress,
			Term:   kv.Term,
		}
		var leaderAcceptedCounter uint64 = 0
		_ = kv.Broadcast(
			"/leader",
			leaderData,
			&leaderAcceptedCounter,
		)
		go kv.heartBeat()

		InfoLogger.Printf("Won election (Term: %d, Yes: %d, No: %d)\n", kv.Term, yesVotes, noVotes)
	} else {
		InfoLogger.Printf("Lost election (Term: %d, Yes: %d, No: %d)\n", kv.Term, yesVotes, noVotes)
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
	success := false
	for retries := 0; retries < MAX_REGISTER_RETRIES; retries++ {
		form := url.Values{}
		form.Add("ip", kv.LocalAddress.String())
		resp, err := http.PostForm(GetURL(entryAddress, "/register"), form)
		if err != nil {
			ErrorLogger.Println(err)
			return nil
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			registrationResponseBytes, _ := ioutil.ReadAll(resp.Body)
			var registrationResponse RegistrationResponseMessage
			err := json.Unmarshal(registrationResponseBytes, &registrationResponse)
			if err != nil {
				ErrorLogger.Println(err)
				ErrorLogger.Printf("Could not parse response body\n")
				return nil
			}

			if registrationResponse.InfoMessage == StatusOKMessage {
				kv.logMutex.Lock()
				jsonContent, _ := json.Marshal(registrationResponse.DatabaseLog)
				InfoLogger.Printf("Got registration response with database log %s\n", jsonContent)
				kv.DatabaseLog = registrationResponse.DatabaseLog

				kv.logMutex.Unlock()
				success = true
				break
			} else {
				ErrorLogger.Println("Unknown registration response not parse response body")
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

		time.Sleep(RETRY_INTERVAL)
	}

	if !success {
		return nil
	}

	// Apply database log
	kv.logMutex.RLock()
	for _, logEntry := range kv.DatabaseLog {
		if logEntry.Committed {
			kv.Database[logEntry.Key] = logEntry.Value
		} else {
			break
		}
	}
	kv.logMutex.RUnlock()

	// Reset last leader heart beat to avoid instant election
	kv.lastLeaderHeartBeat = time.Now()
	go kv.checkLeader()

	return entryAddress
}

//
// Utils
//

func (kv *KeyValueStore) Broadcast(path string, data interface{}, confirmedCounter *uint64) uint64 {
	kv.followerMutex.RLock()
	followerCount := uint64(len(kv.Followers))
	for _, follower := range kv.Followers {
		// Follower is deliberately copied here
		go func(follower Follower, path string, data interface{}, followerCommittedCount *uint64) {
			jsonValue, _ := json.Marshal(data)
			url := GetURL(follower.Address, path)
			for retries := 0; retries < BROADCAST_RETRIES; retries++ {
				resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonValue))
				if err != nil {
					ErrorLogger.Println(err)
					return
				}
				defer resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					break
				}
				time.Sleep(RETRY_INTERVAL)
			}

			atomic.AddUint64(confirmedCounter, 1)
		}(follower, path, data, confirmedCounter)
	}
	kv.followerMutex.RUnlock()

	return followerCount
}
