package model

import (
	"encoding/json"
	"errors"
	"time"
)

type Event struct {
	ID             EventID         `json:"id"`
	Operation      OperationID     `json:"operation"`
	JobID          JobID           `json:"job_id"`
	JobGeneration  uint64          `json:"job_generation"`
	RollGeneration uint64          `json:"roll_generation"`
	Kind           string          `json:"kind"`
	At             time.Time       `json:"at"`
	Payload        json.RawMessage `json:"payload"`
}

func NewEvent(operation OperationID, job Job, kind string, payload any, now time.Time) (Event, error) {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return Event{}, err
	}
	event := Event{ID: NewEventID(), Operation: operation, JobID: job.ID, JobGeneration: job.Generation, RollGeneration: job.RollGeneration, Kind: kind, At: now.UTC(), Payload: encoded}
	return event, event.Validate()
}

func (event Event) Validate() error {
	if event.ID == "" || !event.Operation.Valid() || !event.JobID.Valid() {
		return errors.New("event identity is incomplete")
	}
	if event.JobGeneration == 0 || event.Kind == "" || event.At.IsZero() {
		return errors.New("event metadata is incomplete")
	}
	if !json.Valid(event.Payload) {
		return errors.New("event payload is invalid")
	}
	return nil
}

func DecodeEventPayload[T any](event Event) (T, error) {
	var value T
	err := json.Unmarshal(event.Payload, &value)
	return value, err
}
