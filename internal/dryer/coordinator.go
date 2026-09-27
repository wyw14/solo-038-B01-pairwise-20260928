package dryer

import (
	"errors"
	"fmt"
	"sync"
)

type StartupStatus string

const (
	StartupIdle   StartupStatus = "idle"
	StartupReady  StartupStatus = "ready"
	StartupFailed StartupStatus = "failed"
)

type Coordinator struct {
	mu          sync.RWMutex
	state       *State
	status      StartupStatus
	startupErrs []error
}

func NewCoordinator(state *State) *Coordinator {
	return &Coordinator{state: state, status: StartupIdle}
}

func (coordinator *Coordinator) Startup(damperErr error, folderReady bool) error {
	coordinator.mu.Lock()
	defer coordinator.mu.Unlock()
	coordinator.startupErrs = nil
	if damperErr != nil {
		coordinator.startupErrs = append(coordinator.startupErrs, fmt.Errorf("dryer damper: %w", damperErr))
	}
	if !folderReady {
		coordinator.startupErrs = append(coordinator.startupErrs, errors.New("folder is not ready"))
	}
	if len(coordinator.startupErrs) > 0 {
		coordinator.status = StartupFailed
		return errors.Join(coordinator.startupErrs...)
	}
	coordinator.status = StartupReady
	return nil
}

func (coordinator *Coordinator) Status() (StartupStatus, []string) {
	coordinator.mu.RLock()
	defer coordinator.mu.RUnlock()
	messages := make([]string, 0, len(coordinator.startupErrs))
	for _, err := range coordinator.startupErrs {
		messages = append(messages, err.Error())
	}
	return coordinator.status, messages
}

func (coordinator *Coordinator) State() *State {
	return coordinator.state
}
