package model

import (
	"fmt"
	"sort"
	"time"
)

type CommandKind string

const (
	CommandStop     CommandKind = "stop"
	CommandFeed     CommandKind = "feed"
	CommandHeat     CommandKind = "heat"
	CommandPhase    CommandKind = "phase"
	CommandSplice   CommandKind = "splice"
	CommandInkRoute CommandKind = "ink_route"
)

type Command struct {
	Operation      OperationID `json:"operation"`
	JobID          JobID       `json:"job_id"`
	JobGeneration  uint64      `json:"job_generation"`
	RollGeneration uint64      `json:"roll_generation"`
	Epoch          uint64      `json:"epoch"`
	Kind           CommandKind `json:"kind"`
	IssuedAt       time.Time   `json:"issued_at"`
}

type Acknowledgement struct {
	Unit      string      `json:"unit"`
	Operation OperationID `json:"operation"`
	Epoch     uint64      `json:"epoch"`
	Value     float64     `json:"value"`
	At        time.Time   `json:"at"`
}

func NewCommand(job Job, epoch uint64, kind CommandKind, now time.Time) Command {
	return Command{Operation: NewOperationID(), JobID: job.ID, JobGeneration: job.Generation, RollGeneration: job.RollGeneration, Epoch: epoch, Kind: kind, IssuedAt: now.UTC()}
}

func (command Command) Validate() error {
	if !command.Operation.Valid() || !command.JobID.Valid() {
		return fmt.Errorf("command identities are invalid")
	}
	if command.Epoch == 0 || command.JobGeneration == 0 {
		return fmt.Errorf("command generation and epoch are required")
	}
	if command.Kind == "" {
		return fmt.Errorf("command kind is required")
	}
	return nil
}

func SortedUnits(values map[string]Acknowledgement) []string {
	units := make([]string, 0, len(values))
	for unit := range values {
		units = append(units, unit)
	}
	sort.Strings(units)
	return units
}
