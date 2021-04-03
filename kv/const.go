package kv

import (
	"math/rand"
	"net"
	"time"
)

var BASE_IP_ADDRESS = net.IPv4(172, 23, 0, 0)
var LEADER_IP_ADDRESS = net.IPv4(172, 23, 0, 2)

const PORT string = ":8080"
const MAX_REGISTER_RETRIES = 5

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
