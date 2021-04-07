package kv

import (
	"log"
	"math/rand"
	"net"
	"os"
	"time"
)

var BASE_IP_ADDRESS = net.IPv4(172, 23, 0, 0)
var LEADER_IP_ADDRESS = net.IPv4(172, 23, 0, 2)

const PORT string = ":8080"
const MAX_REGISTER_RETRIES = 5
const BROADCAST_RETRIES = 5
const RETRY_INTERVAL = 10 * time.Millisecond

const max_election_timeout_ms = 1000
const max_election_timeout_diff = 500

// The actual election timeout is randomized between [0.5, 1.0] * MAX_ELECTION_TIMEOUT to reduce the chances of split votes
const MAX_ELECTION_TIMEOUT = max_election_timeout_ms * time.Millisecond

var randomSource = rand.NewSource(time.Now().UnixNano())
var randomNumberGenerator = rand.New(randomSource)

// INDIVIDUAL_ELECTION_TIMEOUT is between 50% and 100% of MAX_ELECTION_TIMEOUT
var INDIVIDUAL_ELECTION_TIMEOUT = MAX_ELECTION_TIMEOUT -
	time.Duration(randomNumberGenerator.Intn(max_election_timeout_diff))*time.Millisecond

var LEADER_HEART_BEAT_TIMEOUT = MAX_ELECTION_TIMEOUT / 3

var INITIAL_LOG = CreateKeyValueLog("initial", "value", false, true)

//
// Logging
//

var (
	InfoLogger  *log.Logger
	ErrorLogger *log.Logger
)

func init() {
	InfoLogger = log.New(os.Stdout, "INFO:  ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}
