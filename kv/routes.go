package kv

import (
	"encoding/json"
	"net/http"
	"os"
)

func status(w http.ResponseWriter, req *http.Request) {
	data := make(map[string]string)
	data["status"] = "ok"

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}

func kill(w http.ResponseWriter, req *http.Request) {
	data := make(map[string]string)
	data["status"] = "ok"

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)

	os.Exit(0)
}

func (kv KeyValueStore) content(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(kv.database)
}
