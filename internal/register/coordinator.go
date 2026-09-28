package register

import (
	"fmt"
	"sync"
	"time"

	"github.com/wyw14/solo-038/internal/journal"
	"github.com/wyw14/solo-038/internal/model"
	"github.com/wyw14/solo-038/internal/sensor"
)

type Coordinator struct {
	mu        sync.Mutex
	state     *State
	journal   *journal.Store
	released  map[model.JobID]journal.IntervalRecord
	lastError error
}

func NewCoordinator(state *State, store *journal.Store) *Coordinator {
	return &Coordinator{state: state, journal: store, released: map[model.JobID]journal.IntervalRecord{}}
}

func (coordinator *Coordinator) AcceptFrame(frame sensor.CameraFrame, maximumGap time.Duration) error {
	window := coordinator.state.Snapshot()
	if !window.Valid {
		return fmt.Errorf("register window is not active")
	}
	if frame.JobGeneration != window.JobGeneration || frame.PlateGeneration != window.PlateGeneration {
		return fmt.Errorf("camera frame does not belong to active register generation")
	}
	if frame.ReceivedAt.Sub(window.LastAt) > maximumGap {
		coordinator.state.Break()
		return fmt.Errorf("camera gap exceeded register dwell")
	}
	return coordinator.state.Add(frame)
}

func (coordinator *Coordinator) Stable(minimum int) bool {
	window := coordinator.state.Snapshot()
	return window.Valid && window.Samples >= minimum
}

func (coordinator *Coordinator) Release(job model.Job, operation model.OperationID, proofs []model.Proof, now time.Time) error {
	window := coordinator.state.Snapshot()
	if !window.Valid || window.Samples == 0 {
		return fmt.Errorf("register window is incomplete")
	}
	interval := model.Interval{JobID: job.ID, JobGeneration: job.Generation, RollGeneration: job.RollGeneration, StartMeter: window.StartMeter, EndMeter: window.LastMeter, OpenedAt: window.StartedAt, ClosedAt: now.UTC()}
	record := journal.IntervalRecord{Interval: interval, Proofs: proofs, Released: false}
	coordinator.mu.Lock()
	defer coordinator.mu.Unlock()
	coordinator.released[job.ID] = record
	if _, err := journal.PersistInterval(coordinator.journal, job, operation, record, now); err != nil {
		coordinator.lastError = err
		return err
	}
	record.Released = true
	coordinator.released[job.ID] = record
	return nil
}

func (coordinator *Coordinator) CanAdvance(job model.Job) bool {
	coordinator.mu.Lock()
	defer coordinator.mu.Unlock()
	record, ok := coordinator.released[job.ID]
	return ok && record.Released && record.Interval.JobGeneration == job.Generation && coordinator.lastError == nil
}

// LotEligibility reports why a job's released interval cannot back a finished
// folder lot, without reserving it.
func (coordinator *Coordinator) LotEligibility(job model.Job) error {
	coordinator.mu.Lock()
	defer coordinator.mu.Unlock()
	return coordinator.lotEligibilityLocked(job)
}

func (coordinator *Coordinator) lotEligibilityLocked(job model.Job) error {
	record, ok := coordinator.released[job.ID]
	if !ok {
		return fmt.Errorf("job has not released a paper interval")
	}
	if !record.Released {
		return fmt.Errorf("released interval is not durable")
	}
	if record.Interval.JobGeneration != job.Generation {
		return fmt.Errorf("released interval does not match current job generation")
	}
	if record.CompletedLot {
		return fmt.Errorf("released interval already produced a folder lot")
	}
	return nil
}

// ConsumeReleased reserves the job's released interval for a single completed
// folder lot. It rejects jobs that never released an interval, intervals from
// another generation, and intervals that already produced a lot.
func (coordinator *Coordinator) ConsumeReleased(job model.Job) error {
	coordinator.mu.Lock()
	defer coordinator.mu.Unlock()
	if err := coordinator.lotEligibilityLocked(job); err != nil {
		return err
	}
	record := coordinator.released[job.ID]
	record.CompletedLot = true
	coordinator.released[job.ID] = record
	return nil
}

// ReleaseConsumedLot reverses a reservation when the folder completion fails
// after the interval was consumed, so the interval stays available for retry.
func (coordinator *Coordinator) ReleaseConsumedLot(job model.Job) {
	coordinator.mu.Lock()
	defer coordinator.mu.Unlock()
	if record, ok := coordinator.released[job.ID]; ok && record.CompletedLot {
		record.CompletedLot = false
		coordinator.released[job.ID] = record
	}
}

func (coordinator *Coordinator) LastError() error {
	coordinator.mu.Lock()
	defer coordinator.mu.Unlock()
	return coordinator.lastError
}
