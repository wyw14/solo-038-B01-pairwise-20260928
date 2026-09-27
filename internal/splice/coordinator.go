package splice

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/wyw14/solo-038/internal/model"
	"github.com/wyw14/solo-038/internal/unwind"
)

type Coordinator struct {
	mu          sync.Mutex
	state       *State
	festoon     *unwind.Festoon
	leases      *LeaseManager
	cancel      context.CancelFunc
	done        chan struct{}
	activity    atomic.Int64
	reservation unwind.Reservation
}

func NewCoordinator(state *State, festoon *unwind.Festoon, leases *LeaseManager) *Coordinator {
	return &Coordinator{state: state, festoon: festoon, leases: leases}
}

func (coordinator *Coordinator) Prepare(owner string, meters float64, beforeCommit func(), now time.Time) (model.OperationID, error) {
	lease, err := coordinator.leases.Acquire(owner, 10*time.Second, now)
	if err != nil {
		return "", err
	}
	operation := model.NewOperationID()
	if beforeCommit != nil {
		beforeCommit()
	}
	reservation, err := coordinator.festoon.Reserve(string(operation), meters)
	if err != nil {
		coordinator.leases.Release(lease)
		return "", err
	}
	if err := coordinator.state.Begin(operation, lease.ID, now); err != nil {
		coordinator.festoon.Release(string(operation), reservation.Fence)
		coordinator.leases.Release(lease)
		return "", err
	}
	coordinator.mu.Lock()
	coordinator.reservation = reservation
	coordinator.mu.Unlock()
	return operation, nil
}

func (coordinator *Coordinator) StartWorker(parent context.Context, operation model.OperationID, interval time.Duration) error {
	coordinator.mu.Lock()
	defer coordinator.mu.Unlock()
	phase, current, _, _, _ := coordinator.state.Snapshot()
	if current != operation || phase != PhasePrepared {
		return fmt.Errorf("splice worker cannot start for inactive operation")
	}
	ctx, cancel := context.WithCancel(parent)
	coordinator.cancel = cancel
	coordinator.done = make(chan struct{})
	if err := coordinator.state.SetWorker(operation, string(model.NewOperationID())); err != nil {
		return err
	}
	go func(done chan struct{}) {
		defer close(done)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				coordinator.activity.Add(1)
			}
		}
	}(coordinator.done)
	return coordinator.state.Advance(operation, PhaseFeeding, time.Now())
}

func (coordinator *Coordinator) Cancel(operation model.OperationID, now time.Time) error {
	coordinator.mu.Lock()
	cancel := coordinator.cancel
	done := coordinator.done
	coordinator.mu.Unlock()
	if err := coordinator.state.Advance(operation, PhaseCancelled, now); err != nil {
		return err
	}
	if cancel != nil {
		cancel()
	}
	if done != nil {
		select {
		case <-done:
		case <-time.After(time.Second):
			return fmt.Errorf("splice worker did not stop")
		}
	}
	return nil
}

func (coordinator *Coordinator) Activity() int64 {
	return coordinator.activity.Load()
}

func (coordinator *Coordinator) FestoonAvailable() float64 {
	return coordinator.festoon.Available()
}
