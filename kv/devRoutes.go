package kv

import (
	"fmt"
	"net"
	"net/http"
	"os"
)

func handleDevKill(w http.ResponseWriter, r *http.Request) {
	RespondJSON(w, http.StatusOK, StatusOKMessage)
	os.Exit(0)
}

func (kv *KeyValueStore) handleDevState(w http.ResponseWriter, r *http.Request) {
	RespondJSON(w, http.StatusOK, StateMessage{
		StatusOKMessage,
		*kv,
	})
}

func (kv *KeyValueStore) handleDevRegister(w http.ResponseWriter, r *http.Request) {
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

	newLeaderAddress := kv.register(address)
	if newLeaderAddress != nil {
		kv.LeaderAddress = newLeaderAddress
		RespondJSON(w, http.StatusOK, StatusOKMessage)
	} else {
		fmt.Println("Registration unsuccessful")
		RespondJSON(w, http.StatusInternalServerError, InfoMessage{
			Status:  "error",
			Message: "Unknown internal server error"})
	}
}
