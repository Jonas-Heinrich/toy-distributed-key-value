package kv

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
)

type KeyValueLog struct {
	Hash      string    `json:"hash"`
	Time      time.Time `json:"time"`
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	Committed bool      `json:"committed"`
}

func CreateKeyValueLog(key string, value string, creationTimeNow bool, commited bool) *KeyValueLog {
	var creationTime time.Time
	if creationTimeNow {
		creationTime = time.Now()
	} else {
		creationTime = time.Unix(0, 0)
	}

	entryHash := sha256.New()
	entryHash.Write([]byte(creationTime.String()))
	entryHash.Write([]byte(key))
	entryHash.Write([]byte(value))

	logEntry := new(KeyValueLog)
	*logEntry = KeyValueLog{
		Hash:      hex.EncodeToString(entryHash.Sum(nil)),
		Time:      creationTime,
		Key:       key,
		Value:     string(value),
		Committed: commited,
	}

	return logEntry
}
