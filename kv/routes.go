package kv

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"reflect"
	"time"

	"github.com/gorilla/mux"
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

	kv.logMutex.RLock()
	kv.followerMutex.Lock()
	kv.Followers = append(
		kv.Followers,
		Follower{
			Address:             address,
			LastLogHash:         INITIAL_LOG.Hash,
			LastCommitedLogHash: INITIAL_LOG.Hash,
		},
	)
	kv.followerMutex.Unlock()
	RespondJSON(w, http.StatusOK, RegistrationResponseMessage{
		InfoMessage: StatusOKMessage,
		DatabaseLog: kv.DatabaseLog,
	})
	jsonContent, _ := json.Marshal(kv.DatabaseLog)
	InfoLogger.Printf("Respond to registration request with database log %s\n", jsonContent)
	kv.logMutex.RUnlock()
}

func (kv *KeyValueStore) handleHeartBeat(w http.ResponseWriter, r *http.Request) {
	if kv.Leader {
		return
	}

	heartBeatMessageBytes, _ := ioutil.ReadAll(r.Body)
	var heartBeatMessage HeartBeatMessage
	if err := json.Unmarshal(heartBeatMessageBytes, &heartBeatMessage); err != nil {
		ErrorLogger.Println("Unspecified hearbeat message format")
		RespondJSON(w, http.StatusInternalServerError, StatusInternalServerErrorMessage)
		return
	}

	kv.Term = heartBeatMessage.Term
	if !reflect.DeepEqual(kv.Followers, heartBeatMessage.Followers) {
		kv.followerMutex.Lock()
		kv.Followers = heartBeatMessage.Followers
		kv.followerMutex.Unlock()
	}
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
		ErrorLogger.Println(err)
		ErrorLogger.Println("Unspecified poll request message format")
		RespondJSON(w, http.StatusInternalServerError, StatusInternalServerErrorMessage)
		return
	}

	if pollRequest.Term >= kv.nextVoteTerm && pollRequest.LastLogHash == kv.DatabaseLog[kv.findLastCommitedLog()].Hash {
		kv.nextVoteTerm = pollRequest.Term + 1
		InfoLogger.Printf("Vote `Yes` (Poll Term: %d, Candidate: %s, Local Term: %d, Next Vote Term: %d)\n", pollRequest.Term, pollRequest.NewLeaderAddress, kv.Term, kv.nextVoteTerm)
		RespondJSON(w, http.StatusOK, PollResponseYes)
	} else {
		InfoLogger.Printf("Vote `No`  (Poll Term: %d, Candidate: %s, Local Term: %d, Next Vote Term: %d)\n", pollRequest.Term, pollRequest.NewLeaderAddress, kv.Term, kv.nextVoteTerm)
		RespondJSON(w, http.StatusOK, PollResponseNo)
	}
}

func (kv *KeyValueStore) handleLeaderUpdate(w http.ResponseWriter, r *http.Request) {
	leaderMessageBytes, _ := ioutil.ReadAll(r.Body)
	var leaderMessage LeaderUpdateMessage
	if err := json.Unmarshal(leaderMessageBytes, &leaderMessage); err != nil {
		ErrorLogger.Println("Unspecified leader update message format")
		os.Exit(1)
	}

	kv.Leader = false
	kv.LeaderAddress = leaderMessage.Leader
	kv.Term = leaderMessage.Term
	kv.lastLeaderHeartBeat = time.Now()
	RespondJSON(w, http.StatusOK, StatusOKMessage)

	InfoLogger.Printf("Accepted new leader (%s)\n", kv.LeaderAddress.String())
}

func (kv *KeyValueStore) handleLeaderRequest(w http.ResponseWriter, r *http.Request) {
	InfoLogger.Println("Returning leader address")
	RespondJSON(w, http.StatusOK, IPMessage{
		InfoMessage: StatusOKMessage,
		IP:          kv.LeaderAddress,
	})
}

//
// Read
//

func (kv *KeyValueStore) handleRead(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	if kv.Leader {
		kv.databaseMutex.RLock()
		value, ok := kv.Database[key]
		kv.databaseMutex.RUnlock()
		if ok {
			RespondJSON(w, http.StatusOK, ReadMessage{
				InfoMessage: StatusOKMessage,
				Value:       value,
			})
		} else {
			RespondJSON(w, http.StatusNotFound, ReadMessage{
				InfoMessage: StatusValueNotFoundMessage,
				Value:       "",
			})
		}
	} else {
		InfoLogger.Println("Proxying read request to leader")
		proxyResp, err := http.Get(GetURL(kv.LeaderAddress, "/read/"+key))
		if err != nil {
			ErrorLogger.Println(err)
			RespondJSON(w, http.StatusInternalServerError, StatusInternalServerErrorMessage)
			return
		}
		defer proxyResp.Body.Close()

		readMessageBytes, _ := ioutil.ReadAll(proxyResp.Body)
		var readMessage ReadMessage
		if err := json.Unmarshal(readMessageBytes, &readMessage); err != nil {
			ErrorLogger.Println("Unspecified read message format")
			RespondJSON(w, http.StatusInternalServerError, StatusInternalServerErrorMessage)
			return
		}

		RespondJSON(w, proxyResp.StatusCode, readMessage)
		InfoLogger.Println("Proxy read request finished and successful")
	}
}

//
// Write
//

func (kv *KeyValueStore) findLastCommitedLog() int {
	// kv.logMutex.RLock()
	for i := len(kv.DatabaseLog) - 1; i >= 0; i-- {
		if kv.DatabaseLog[i].Committed {
			// kv.logMutex.RUnlock()
			return i
		}
	}
	// Should never happen, since log is always initialized with a commited entry
	os.Exit(1)
	return 0
}

func (kv *KeyValueStore) handleLogAppend(w http.ResponseWriter, r *http.Request) {
	logBytes, _ := ioutil.ReadAll(r.Body)
	var logMessages AppendEntriesMessage
	if err := json.Unmarshal(logBytes, &logMessages); err != nil {
		ErrorLogger.Println("Unspecified key value log message format")
		RespondJSON(w, http.StatusInternalServerError, StatusInternalServerErrorMessage)
		return
	}

	if len(logMessages.KeyValueLog) < 2 {
		ErrorLogger.Println("Not enough log messages provided")
		RespondJSON(w, http.StatusInternalServerError, StatusInternalServerErrorMessage)
		return
	} else if !logMessages.KeyValueLog[0].Committed {
		ErrorLogger.Println("First log message should already be committed")
		RespondJSON(w, http.StatusInternalServerError, StatusInternalServerErrorMessage)
		return
	}

	// Note: Since the log is append only, it is safe to determine the startIndex here
	//       and give away the lock afterwards.
	// kv.logMutex.RLock()
	startIndex := 0
	found := false
	for i := len(kv.DatabaseLog) - 1; i >= 0; i-- {
		if kv.DatabaseLog[i].Hash == logMessages.KeyValueLog[0].Hash {
			startIndex = i
			found = true
		}
	}
	// kv.logMutex.RUnlock()

	if !found {
		ErrorLogger.Println("Unknown log reference point")
		RespondJSON(w, http.StatusInternalServerError, StatusInternalServerErrorMessage)
		return
	}

	kv.logMutex.Lock()
	for index, logMessage := range logMessages.KeyValueLog {
		if startIndex+index < len(kv.DatabaseLog) {
			if kv.DatabaseLog[startIndex+index].Hash != logMessage.Hash {
				kv.logMutex.Unlock()
				ErrorLogger.Println("Hash does not match")
				RespondJSON(w, http.StatusInternalServerError, StatusInternalServerErrorMessage)
				return
			}
		} else {
			kv.DatabaseLog = append(kv.DatabaseLog, logMessage)
		}
	}
	kv.logMutex.Unlock()
	RespondJSON(w, http.StatusOK, StatusOKMessage)
	InfoLogger.Printf("Appended up to log %s\n", logMessages.KeyValueLog[len(logMessages.KeyValueLog)-1].Hash)
}

func (kv *KeyValueStore) handleCommit(w http.ResponseWriter, r *http.Request) {
	commitLogBytes, _ := ioutil.ReadAll(r.Body)
	var commitLogMessage CommitLogMessage
	if err := json.Unmarshal(commitLogBytes, &commitLogMessage); err != nil {
		ErrorLogger.Println("Unspecified commit log message format")
		RespondJSON(w, http.StatusInternalServerError, StatusInternalServerErrorMessage)
		return
	}

	kv.logMutex.Lock()

	// Determine both the index for the log that is to be committed..

	var endLogIndex int
	for endLogIndex = len(kv.DatabaseLog) - 1; endLogIndex >= 0; endLogIndex-- {
		if commitLogMessage.LogHash == kv.DatabaseLog[endLogIndex].Hash {
			break
		}
	}

	if endLogIndex == 0 {
		RespondJSON(w, http.StatusNotFound, StatusLogNotFoundMessage)
		return
	}

	// ..and index for the first uncommited log
	beginLogIndex := 0
	for ; kv.DatabaseLog[beginLogIndex].Committed && beginLogIndex < endLogIndex; beginLogIndex++ {
	}

	InfoLogger.Printf("Committing from %d to %d", beginLogIndex, endLogIndex)

	kv.databaseMutex.Lock()
	for i := beginLogIndex; i <= endLogIndex; i++ {
		kv.Database[kv.DatabaseLog[i].Key] = kv.DatabaseLog[i].Value
		kv.DatabaseLog[i].Committed = true
	}
	kv.databaseMutex.Unlock()

	kv.logMutex.Unlock()

	RespondJSON(w, http.StatusOK, StatusOKMessage)
	InfoLogger.Printf("Commited up to log %s\n", commitLogMessage.LogHash)
}

func (kv *KeyValueStore) distributeChange(logEntry *KeyValueLog, appendedLogIndex int) {
	//
	// Append from last known log
	//

	InfoLogger.Printf("Appending log %s", logEntry.Hash)

	lastCommitIndex := kv.findLastCommitedLog()
	appendData := &AppendEntriesMessage{KeyValueLog: kv.DatabaseLog[lastCommitIndex : appendedLogIndex+1]}
	var appendedCounter uint64 = 0
	followerCount := kv.Broadcast(
		"/log/append",
		appendData,
		&appendedCounter,
	)

	majorityCount := uint64(float32(followerCount)*0.5) + 1 // Half plus one
	for appendedCounter < majorityCount {
		time.Sleep(100 * time.Microsecond)
	}

	InfoLogger.Printf("Log %s considered appended", logEntry.Hash)

	//
	// Commit new log
	//

	InfoLogger.Printf("Commiting log %s", logEntry.Hash)
	commitData := &CommitLogMessage{LogHash: logEntry.Hash}
	var committedCounter uint64 = 0
	followerCount = kv.Broadcast(
		"/log/commit",
		commitData,
		&committedCounter,
	)

	majorityCount = uint64(float32(followerCount)*0.5) + 1 // Half plus one
	for committedCounter < majorityCount {
		time.Sleep(100 * time.Microsecond)
	}

	kv.databaseMutex.Lock()
	kv.Database[logEntry.Key] = logEntry.Value
	kv.databaseMutex.Unlock()
	logEntry.Committed = true
	InfoLogger.Printf("Log %s is now considered committed", logEntry.Hash)
}

func (kv *KeyValueStore) handleWrite(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	if kv.Leader {
		value, _ := ioutil.ReadAll(r.Body)
		logEntry := CreateKeyValueLog(key, string(value), true, false)

		kv.logMutex.Lock()
		kv.DatabaseLog = append(kv.DatabaseLog, logEntry)
		appendedLogIndex := len(kv.DatabaseLog) - 1
		kv.logMutex.Unlock()

		kv.distributeChange(logEntry, appendedLogIndex)
		RespondJSON(w, http.StatusOK, StatusOKMessage)
		return
	} else {
		proxyResp, err := http.Post(GetURL(kv.LeaderAddress, "/write/"+key), "application/json", r.Body)
		if err != nil {
			ErrorLogger.Println(err)
			RespondJSON(w, http.StatusInternalServerError, StatusInternalServerErrorMessage)
			return
		}
		defer proxyResp.Body.Close()

		infoMessageBytes, _ := ioutil.ReadAll(proxyResp.Body)
		var infoMessage InfoMessage
		if err := json.Unmarshal(infoMessageBytes, &infoMessage); err != nil {
			ErrorLogger.Println("Unspecified info message format")
			RespondJSON(w, http.StatusInternalServerError, StatusInternalServerErrorMessage)
			return
		}

		RespondJSON(w, proxyResp.StatusCode, infoMessage)
		return
	}
}
