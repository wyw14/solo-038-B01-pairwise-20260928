package job

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/wyw14/solo-038/internal/model"
)

type State struct {
	mu      sync.RWMutex
	jobs    map[model.JobID]model.Job
	current model.JobID
}

func NewState() *State {
	return &State{jobs: map[model.JobID]model.Job{}}
}

func (state *State) Create(now time.Time) model.Job {
	state.mu.Lock()
	defer state.mu.Unlock()
	created := model.NewJob(now)
	state.jobs[created.ID] = created
	state.current = created.ID
	return created
}

func (state *State) Put(value model.Job) error {
	if err := value.Validate(); err != nil {
		return err
	}
	state.mu.Lock()
	defer state.mu.Unlock()
	if current, ok := state.jobs[value.ID]; ok && value.Generation < current.Generation {
		return fmt.Errorf("job generation moved backwards")
	}
	state.jobs[value.ID] = value
	state.current = value.ID
	return nil
}

func (state *State) Current() (model.Job, bool) {
	state.mu.RLock()
	defer state.mu.RUnlock()
	value, ok := state.jobs[state.current]
	return value, ok
}

func (state *State) Get(id model.JobID) (model.Job, bool) {
	state.mu.RLock()
	defer state.mu.RUnlock()
	value, ok := state.jobs[id]
	return value, ok
}

func (state *State) IDs() []string {
	state.mu.RLock()
	defer state.mu.RUnlock()
	values := make([]string, 0, len(state.jobs))
	for id := range state.jobs {
		values = append(values, string(id))
	}
	sort.Strings(values)
	return values
}
