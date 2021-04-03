package kv

import (
	"net/http"
	"os"
)

func kill(w http.ResponseWriter, req *http.Request) {
	RespondJSON(w, http.StatusOK, StatusOKResponse)
	os.Exit(0)
}

func (kv KeyValueStore) state(w http.ResponseWriter, req *http.Request) {
	RespondJSON(w, http.StatusOK, StateResponse{
		StatusOKResponse,
		kv,
	})
}
