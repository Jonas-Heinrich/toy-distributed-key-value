package kv

import (
	"fmt"
	"net"
	"net/http"
	"os"
)

func handleDevKill(w http.ResponseWriter, r *http.Request) {
	RespondJSON(w, http.StatusOK, StatusOKResponse)
	os.Exit(0)
}

func (kv *KeyValueStore) handleDevState(w http.ResponseWriter, r *http.Request) {
	RespondJSON(w, http.StatusOK, StateResponse{
		StatusOKResponse,
		*kv,
	})
}

func (kv *KeyValueStore) handleDevRegister(w http.ResponseWriter, r *http.Request) {
	rawAddress := r.FormValue("ip")
	if rawAddress == "" {
		RespondJSON(w, http.StatusBadRequest, MessageResponse{
			Status:  "error",
			Message: "No ip address provided"})
		return
	}

	address := net.ParseIP(rawAddress)
	if address == nil {
		RespondJSON(w, http.StatusBadRequest, MessageResponse{
			Status:  "error",
			Message: "IP does not match expected format"})
		return
	}

	newLeaderAddress := kv.register(address)
	if newLeaderAddress != nil {
		kv.LeaderAddress = newLeaderAddress
		RespondJSON(w, http.StatusOK, StatusOKResponse)
	} else {
		fmt.Println("Registration unsuccessful")
		RespondJSON(w, http.StatusInternalServerError, MessageResponse{
			Status:  "error",
			Message: "Unknown internal server error"})
	}
}
