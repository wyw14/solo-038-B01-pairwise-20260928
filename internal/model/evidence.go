package model

import (
	"errors"
	"sort"
	"time"
)

type Interval struct {
	JobID          JobID     `json:"job_id"`
	JobGeneration  uint64    `json:"job_generation"`
	RollGeneration uint64    `json:"roll_generation"`
	StartMeter     float64   `json:"start_meter"`
	EndMeter       float64   `json:"end_meter"`
	OpenedAt       time.Time `json:"opened_at"`
	ClosedAt       time.Time `json:"closed_at"`
}

type Proof struct {
	Kind           string    `json:"kind"`
	Source         string    `json:"source"`
	JobGeneration  uint64    `json:"job_generation"`
	RollGeneration uint64    `json:"roll_generation"`
	ValidFrom      float64   `json:"valid_from"`
	ValidTo        float64   `json:"valid_to"`
	ObservedAt     time.Time `json:"observed_at"`
}

type ProofSet struct {
	Interval Interval `json:"interval"`
	Proofs   []Proof  `json:"proofs"`
}

func (interval Interval) Validate() error {
	if !interval.JobID.Valid() || interval.JobGeneration == 0 || interval.RollGeneration == 0 {
		return errors.New("interval identity is incomplete")
	}
	if interval.EndMeter <= interval.StartMeter {
		return errors.New("interval meters are invalid")
	}
	if interval.ClosedAt.Before(interval.OpenedAt) {
		return errors.New("interval time range is invalid")
	}
	return nil
}

func (set ProofSet) Complete(required ...string) bool {
	wanted := make(map[string]bool, len(required))
	for _, kind := range required {
		wanted[kind] = false
	}
	for _, proof := range set.Proofs {
		if proof.JobGeneration != set.Interval.JobGeneration || proof.RollGeneration != set.Interval.RollGeneration {
			continue
		}
		if proof.ValidFrom > set.Interval.StartMeter || proof.ValidTo < set.Interval.EndMeter {
			continue
		}
		if _, ok := wanted[proof.Kind]; ok {
			wanted[proof.Kind] = true
		}
	}
	for _, present := range wanted {
		if !present {
			return false
		}
	}
	return true
}

func (set ProofSet) Sources() []string {
	seen := map[string]bool{}
	for _, proof := range set.Proofs {
		seen[proof.Source] = true
	}
	values := make([]string, 0, len(seen))
	for source := range seen {
		values = append(values, source)
	}
	sort.Strings(values)
	return values
}
