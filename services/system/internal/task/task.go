// Package task defines system-owned, versioned queue contracts. No user token is serialized.
package task

import (
	"encoding/json"
	"errors"
	"github.com/hibiken/asynq"
	"time"
)

const Execute = "system.execute.v1"
const Tick = "system.cleanup.tick.v1"
const Export = "system.dictionary_export.v1"
const Delete = "system.file_delete.v1"
const Cleanup = "system.upload_cleanup.v1"

type Payload struct {
	ID         uint64 `json:"id"`
	Generation int64  `json:"generation"`
}
type Schedule struct {
	Slot int64 `json:"slot"`
}

func Message(id uint64, generation int64) *asynq.Task {
	p, _ := json.Marshal(Payload{id, generation})
	return asynq.NewTask(Execute, p)
}
func Decode(data []byte) (Payload, error) {
	var p Payload
	err := json.Unmarshal(data, &p)
	if err != nil || p.ID == 0 || p.Generation < 1 {
		return p, errors.New("invalid task payload")
	}
	return p, nil
}
func TickMessage(now time.Time) *asynq.Task {
	p, _ := json.Marshal(Schedule{now.UTC().Truncate(time.Hour).Unix()})
	return asynq.NewTask(Tick, p)
}
