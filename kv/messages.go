package kv

import "net"

type InfoMessage struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

var StatusOKMessage = InfoMessage{"OK", "OK"}
var StatusMovedMessage = InfoMessage{"not responsible", "Node is not responsible, use provided address"}
var StatusMissingURLParameterMessage = InfoMessage{"URL parameter missing", "A required URL parameter seems to be missing"}
var StatusBadURLParameterMessage = InfoMessage{"URL parameter malformed", "A URL parameter does not match its specification (count, form, ..)"}
var StatusInternalServerErrorMessage = InfoMessage{"error occurred", "An unknown internal server error appeared"}

type StateMessage struct {
	InfoMessage   InfoMessage
	KeyValueStore KeyValueStore
}

type IPMessage struct {
	InfoMessage InfoMessage
	IP          net.IP `json:"ip"`
}

type HeartBeatMessage struct {
	InfoMessage       InfoMessage
	Term              int
	FollowerAddresses []net.IP `json:"followerAddresses"`
}

type PollRequestMessage struct {
	Term             int    `json:"term"`
	NewLeaderAddress net.IP `json:"newLeaderAddress"`
}

type PollResponseMessage struct {
	Yes bool `json:"vote"`
}

var PollResponseYes = PollResponseMessage{Yes: true}
var PollResponseNo = PollResponseMessage{Yes: false}

type LeaderUpdateMessage struct {
	Leader net.IP `json:"leader"`
	Term   int    `json:"term"`
}

//
// Read/Write
//

type ReadMessage struct {
	InfoMessage InfoMessage
	Value       string `json:"value"`
}

var StatusValueNotFoundMessage = InfoMessage{"Value not found", "The requested key could not be found in the database"}
