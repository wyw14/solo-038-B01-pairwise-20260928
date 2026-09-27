package splice

import (
	"fmt"
	"sync"
	"time"

	"github.com/wyw14/solo-038/internal/model"
)

type Phase string

const (
	PhaseIdle      Phase = "idle"
	PhasePrepared  Phase = "prepared"
	PhaseFeeding   Phase = "feeding"
	PhaseBonding   Phase = "bonding"
	PhaseCancelled Phase = "cancelled"
	PhaseComplete  Phase = "complete"
)

type State struct {
	mu        sync.RWMutex
	phase     Phase
	operation model.OperationID
	lease     model.LeaseID
	worker    string
	updatedAt time.Time
}

func NewState() *State {
	return &State{phase: PhaseIdle}
}

func (state *State) Begin(operation model.OperationID, lease model.LeaseID, now time.Time) error {
	if !operation.Valid() || !lease.Valid() {
		return fmt.Errorf("splice operation and lease are required")
	}
	state.mu.Lock()
	defer state.mu.Unlock()
	if state.phase != PhaseIdle && state.phase != PhaseComplete && state.phase != PhaseCancelled {
		return fmt.Errorf("splice operation is already active")
	}
	state.operation = operation
	state.lease = lease
	state.phase = PhasePrepared
	state.updatedAt = now.UTC()
	return nil
}

func (state *State) Advance(operation model.OperationID, next Phase, now time.Time) error {
	state.mu.Lock()
	defer state.mu.Unlock()
	if operation != state.operation {
		return fmt.Errorf("splice operation fence rejected")
	}
	allowed := map[Phase]map[Phase]bool{PhasePrepared: {PhaseFeeding: true, PhaseCancelled: true}, PhaseFeeding: {PhaseBonding: true, PhaseCancelled: true}, PhaseBonding: {PhaseComplete: true, PhaseCancelled: true}}
	if !allowed[state.phase][next] {
		return fmt.Errorf("splice transition from %s to %s is invalid", state.phase, next)
	}
	state.phase = next
	state.updatedAt = now.UTC()
	return nil
}

func (state *State) SetWorker(operation model.OperationID, worker string) error {
	state.mu.Lock()
	defer state.mu.Unlock()
	if operation != state.operation || worker == "" {
		return fmt.Errorf("splice worker fence rejected")
	}
	state.worker = worker
	return nil
}

func (state *State) Snapshot() (Phase, model.OperationID, model.LeaseID, string, time.Time) {
	state.mu.RLock()
	defer state.mu.RUnlock()
	return state.phase, state.operation, state.lease, state.worker, state.updatedAt
}
