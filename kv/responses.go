package kv

import "net"

type MessageResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

var StatusOKResponse = MessageResponse{"OK", "OK"}
var StatusMovedResponse = MessageResponse{"not responsible", "Node is not responsible, use provided address"}

type StateResponse struct {
	MessageResponse MessageResponse
	KeyValueStore   KeyValueStore
}

type IPResponse struct {
	MessageResponse MessageResponse
	IP              net.IP `json:"ip"`
}
