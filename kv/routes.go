package kv

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"time"
)

func handleStatus(w http.ResponseWriter, r *http.Request) {
	RespondJSON(w, http.StatusOK, StatusOKMessage)
}

func (kv *KeyValueStore) handleRegister(w http.ResponseWriter, r *http.Request) {
	if !kv.Leader {
		RespondJSON(w, http.StatusServiceUnavailable, IPMessage{
			InfoMessage: StatusMovedMessage,
			IP:          kv.LeaderAddress,
		})
		return
	}

	rawAddress := r.FormValue("ip")
	if rawAddress == "" {
		RespondJSON(w, http.StatusBadRequest, InfoMessage{
			Status:  "error",
			Message: "No ip address provided"})
		return
	}

	address := net.ParseIP(rawAddress)
	if address == nil {
		RespondJSON(w, http.StatusBadRequest, InfoMessage{
			Status:  "error",
			Message: "IP does not match expected format"})
		return
	}

	kv.FollowerAddresses = append(kv.FollowerAddresses, address)
	RespondJSON(w, http.StatusOK, StatusOKMessage)
}

func (kv *KeyValueStore) handleHeartBeat(w http.ResponseWriter, r *http.Request) {
	if kv.Leader {
		return
	}

	heartBeatMessageBytes, _ := ioutil.ReadAll(r.Body)
	var heartBeatMessage HeartBeatMessage
	if err := json.Unmarshal(heartBeatMessageBytes, &heartBeatMessage); err != nil {
		fmt.Println("Unspecified hearbeat message format")
		os.Exit(1)
	}

	kv.Term = heartBeatMessage.Term
	kv.FollowerAddresses = heartBeatMessage.FollowerAddresses
	kv.lastLeaderHeartBeat = time.Now()
	RespondJSON(w, http.StatusOK, StatusOKMessage)
}

//
// Election Handling
//

func (kv *KeyValueStore) handlePoll(w http.ResponseWriter, r *http.Request) {
	if kv.Leader {
		RespondJSON(w, http.StatusOK, PollResponseNo)
		return
	}

	pollParameters, ok := r.URL.Query()["poll_parameters"]
	if !ok {
		RespondJSON(w, http.StatusBadRequest, StatusMissingURLParameterMessage)
		return
	}

	if len(pollParameters) != 1 {
		RespondJSON(w, http.StatusBadRequest, StatusBadURLParameterMessage)
		return
	}

	var pollRequest PollRequestMessage
	if err := json.Unmarshal([]byte(pollParameters[0]), &pollRequest); err != nil {
		fmt.Println(err)
		fmt.Println("Unspecified poll request message format")
		os.Exit(1)
	}

	if pollRequest.Term >= kv.nextVoteTerm {
		kv.nextVoteTerm = pollRequest.Term + 1
		RespondJSON(w, http.StatusOK, PollResponseYes)
	} else {
		RespondJSON(w, http.StatusOK, PollResponseNo)
	}
}

func (kv *KeyValueStore) handleLeaderUpdate(w http.ResponseWriter, r *http.Request) {
	leaderMessageBytes, _ := ioutil.ReadAll(r.Body)
	var leaderMessage LeaderUpdateMessage
	if err := json.Unmarshal(leaderMessageBytes, &leaderMessage); err != nil {
		fmt.Println("Unspecified leader update message format")
		os.Exit(1)
	}

	kv.Leader = false
	kv.LeaderAddress = leaderMessage.Leader
	kv.Term = leaderMessage.Term
	kv.lastLeaderHeartBeat = time.Now()
	RespondJSON(w, http.StatusOK, StatusOKMessage)

	fmt.Printf("Accepted new leader (%s)\n", kv.LeaderAddress.String())
}

func (kv *KeyValueStore) handleLeaderRequest(w http.ResponseWriter, r *http.Request) {
	RespondJSON(w, http.StatusOK, IPMessage{
		InfoMessage: StatusOKMessage,
		IP:          kv.LeaderAddress,
	})
}
