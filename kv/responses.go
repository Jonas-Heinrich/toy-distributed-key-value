package kv

type MessageResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

var StatusOKResponse = MessageResponse{"OK", "OK"}

type StateResponse struct {
	MessageResponse MessageResponse
	KeyValueStore   KeyValueStore
}
