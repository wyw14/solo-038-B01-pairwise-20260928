package register

import (
	"fmt"
	"time"

	"github.com/wyw14/solo-038/internal/journal"
	"github.com/wyw14/solo-038/internal/model"
)

func ReleaseWithProofs(coordinator *Coordinator, job model.Job, operation model.OperationID, proofs []model.Proof, dryerReady bool, now time.Time) error {
	if !dryerReady {
		return fmt.Errorf("dryer proof is not ready")
	}
	if err := coordinator.Release(job, operation, proofs, now); err != nil {
		return err
	}
	return nil
}

func RebuildFromJournal(store *journal.Store) (journal.RecoveredState, error) {
	events, err := store.Events()
	if err != nil {
		return journal.RecoveredState{}, err
	}
	return journal.Replay(events)
}
