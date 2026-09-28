package service

import (
	"fmt"
	"strings"

	"github.com/wyw14/solo-038/internal/model"
)

// FinishGateError reports every production prerequisite that a finish_output
// attempt is missing, so rejected operations retain an auditable reason.
type FinishGateError struct {
	Reasons []string
}

func (err *FinishGateError) Error() string {
	return "finish output rejected: " + strings.Join(err.Reasons, "; ")
}

// requireProductionEvidence ensures paper has actually been produced for the
// active job generation before the folder is allowed to complete a lot:
// a roll must be loaded (fenced), the web threaded and fed, and at least one
// paper interval must have been released with dryer proof.
func (runtime *Runtime) requireProductionEvidence(job model.Job) error {
	var reasons []string
	owner, jobGeneration, rollGeneration, threaded, meter := runtime.web.Snapshot()
	switch {
	case owner == "":
		reasons = append(reasons, "no roll has been loaded for the active job")
	case owner != string(job.ID):
		reasons = append(reasons, fmt.Sprintf("web fence owner %q does not match active job %q", owner, job.ID))
	default:
		if jobGeneration != job.Generation || rollGeneration != job.RollGeneration {
			reasons = append(reasons, "web fence does not match the active job and roll generation")
		}
		if !threaded {
			reasons = append(reasons, "paper web has not been threaded")
		}
		if meter <= 0 {
			reasons = append(reasons, "paper web has not been fed through the press")
		}
	}
	if !runtime.registerControl.CanAdvance(job) {
		reasons = append(reasons, "no paper interval with dryer proof has been released for the active job")
	}
	if len(reasons) > 0 {
		return &FinishGateError{Reasons: reasons}
	}
	return nil
}
