package kv

import (
	"net/http"
)

func status(w http.ResponseWriter, req *http.Request) {
	RespondJSON(w, http.StatusOK, StatusOKResponse)
}
