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

type RegistrationResponseMessage struct {
	InfoMessage InfoMessage
	DatabaseLog []*KeyValueLog `json:"databaseLog"`
}

type StateMessage struct {
	InfoMessage   InfoMessage
	KeyValueStore KeyValueStore
}

type IPMessage struct {
	InfoMessage InfoMessage
	IP          net.IP `json:"ip"`
}

type HeartBeatMessage struct {
	InfoMessage InfoMessage
	Term        uint64     `json:"term"`
	Followers   []Follower `json:"followers"`
}

type PollRequestMessage struct {
	Term             uint64 `json:"term"`
	NewLeaderAddress net.IP `json:"newLeaderAddress"`
	LastLogHash      string `json:"lastLogHash"`
}

type PollResponseMessage struct {
	Yes bool `json:"vote"`
}

var PollResponseYes = PollResponseMessage{Yes: true}
var PollResponseNo = PollResponseMessage{Yes: false}

type LeaderUpdateMessage struct {
	Leader net.IP `json:"leader"`
	Term   uint64 `json:"term"`
}

//
// Read
//

type ReadMessage struct {
	InfoMessage InfoMessage
	Value       string `json:"value"`
}

var StatusValueNotFoundMessage = InfoMessage{"Value not found", "The requested key could not be found in the database"}

//
// Write
//

type AppendEntriesMessage struct {
	KeyValueLog []*KeyValueLog `json:"logs"`
}

type CommitLogMessage struct {
	LogHash string `json:"logHash"`
}

var StatusLogNotFoundMessage = InfoMessage{"Log not found", "The requested log could not be found in the database log, it remains uncommited."}
