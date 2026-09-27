package job

import (
	"fmt"

	"github.com/wyw14/solo-038/internal/journal"
	"github.com/wyw14/solo-038/internal/model"
)

type RecoveryDecision struct {
	Job            model.Job `json:"job"`
	MayFeed        bool      `json:"may_feed"`
	MayRelease     bool      `json:"may_release"`
	RecoveredOwner string    `json:"recovered_owner"`
}

func DecideRecovery(recovered journal.RecoveredState) (RecoveryDecision, error) {
	if err := recovered.Job.Validate(); err != nil {
		return RecoveryDecision{}, err
	}
	decision := RecoveryDecision{Job: recovered.Job, RecoveredOwner: recovered.Owner}
	decision.MayFeed = recovered.Owner != "" && recovered.Job.RollGeneration > 0 && recovered.Job.State != model.JobCancelled && recovered.Job.State != model.JobComplete
	decision.MayRelease = len(recovered.ReleasedRanges) > 0
	if decision.MayFeed && recovered.Owner == "" {
		return RecoveryDecision{}, fmt.Errorf("feed recovery has no owner fence")
	}
	return decision, nil
}

func RestoreState(state *State, decision RecoveryDecision) error {
	if err := decision.Job.Validate(); err != nil {
		return err
	}
	return state.Put(decision.Job)
}
