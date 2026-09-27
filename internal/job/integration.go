package job

import (
	"fmt"
	"time"

	"github.com/wyw14/solo-038/internal/journal"
	"github.com/wyw14/solo-038/internal/model"
)

func Transition(state *State, store *journal.Store, job model.Job, next model.JobState, now time.Time) (model.Job, error) {
	updated, err := job.Transition(next, now)
	if err != nil {
		return job, err
	}
	operation := model.NewOperationID()
	if _, err := journal.PersistTransition(store, updated, operation, job.State, next, now); err != nil {
		return job, err
	}
	if err := state.Put(updated); err != nil {
		return job, err
	}
	return updated, nil
}

func RequireSpliceReady(job model.Job, pressStopped, webFenced, leaseValid bool) error {
	if job.State != model.JobPrinting && job.State != model.JobSplicing {
		return fmt.Errorf("job state does not permit splicing")
	}
	if !pressStopped || !webFenced || !leaseValid {
		return fmt.Errorf("splice interlocks are incomplete")
	}
	return nil
}

func CurrentGeneration(state *State) (uint64, error) {
	current, ok := state.Current()
	if !ok {
		return 0, fmt.Errorf("there is no current job")
	}
	return current.Generation, nil
}
