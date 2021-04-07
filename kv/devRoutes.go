package kv

import (
	"net"
	"net/http"
	"os"
)

func handleDevKill(w http.ResponseWriter, r *http.Request) {
	RespondJSON(w, http.StatusOK, StatusOKMessage)
	os.Exit(0)
}

func (kv *KeyValueStore) handleDevState(w http.ResponseWriter, r *http.Request) {
	kv.databaseMutex.RLock()
	kv.logMutex.RLock()
	kv.followerMutex.RLock()
	RespondJSON(w, http.StatusOK, StateMessage{
		StatusOKMessage,
		*kv,
	})
	kv.databaseMutex.RUnlock()
	kv.logMutex.RUnlock()
	kv.followerMutex.RUnlock()
}

func (kv *KeyValueStore) handleDevRegister(w http.ResponseWriter, r *http.Request) {
	rawAddress := r.FormValue("ip")
	if rawAddress == "" {
		ErrorLogger.Println("No ip address provided")
		RespondJSON(w, http.StatusBadRequest, InfoMessage{
			Status:  "error",
			Message: "No ip address provided"})
		return
	}

	address := net.ParseIP(rawAddress)
	if address == nil {
		ErrorLogger.Println("IP does not match expected format")
		RespondJSON(w, http.StatusBadRequest, InfoMessage{
			Status:  "error",
			Message: "IP does not match expected format"})
		return
	}

	newLeaderAddress := kv.register(address)
	if newLeaderAddress != nil {
		kv.LeaderAddress = newLeaderAddress
		RespondJSON(w, http.StatusOK, StatusOKMessage)
	} else {
		ErrorLogger.Println("Registration unsuccessful")
		RespondJSON(w, http.StatusInternalServerError, InfoMessage{
			Status:  "error",
			Message: "Unknown internal server error"})
	}
}
