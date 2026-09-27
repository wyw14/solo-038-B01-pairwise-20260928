package folder

import "sync"

type State struct {
	mu            sync.RWMutex
	ready         bool
	speed         float64
	cutLength     float64
	lastFault     string
	completedLots int
}

func NewState() *State {
	return &State{}
}

func (state *State) SetReady(ready bool, fault string) {
	state.mu.Lock()
	defer state.mu.Unlock()
	state.ready = ready
	state.lastFault = fault
}

func (state *State) Configure(speed, cutLength float64) bool {
	state.mu.Lock()
	defer state.mu.Unlock()
	if speed <= 0 || cutLength <= 0 {
		return false
	}
	state.speed = speed
	state.cutLength = cutLength
	return true
}

func (state *State) CompleteLot() bool {
	state.mu.Lock()
	defer state.mu.Unlock()
	if !state.ready {
		return false
	}
	state.completedLots++
	return true
}

func (state *State) Snapshot() (bool, float64, float64, string, int) {
	state.mu.RLock()
	defer state.mu.RUnlock()
	return state.ready, state.speed, state.cutLength, state.lastFault, state.completedLots
}
