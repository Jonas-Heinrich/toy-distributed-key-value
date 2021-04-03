package kv

import (
	"net"
	"net/http"
)

func handleStatus(w http.ResponseWriter, r *http.Request) {
	RespondJSON(w, http.StatusOK, StatusOKResponse)
}

func (kv *KeyValueStore) handleRegister(w http.ResponseWriter, r *http.Request) {
	if !kv.Leader {
		RespondJSON(w, http.StatusServiceUnavailable, IPResponse{
			MessageResponse: StatusMovedResponse,
			IP:              kv.LeaderAddress,
		})
		return
	}

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

	kv.FollowerAddresses = append(kv.FollowerAddresses, address)
	RespondJSON(w, http.StatusOK, StatusOKResponse)
}
