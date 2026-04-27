package model

import (
	"encoding/json"
	"time"
)

type IdempotencyRecord struct {
	Key         string          `json:"key"`
	RequestHash string          `json:"request_hash"`
	Response    json.RawMessage `json:"response"`
	CreatedAt   time.Time       `json:"created_at"`
}
