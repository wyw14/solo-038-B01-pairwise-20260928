package ink

import (
	"sort"
	"sync"
)

type CircuitState string

const (
	CircuitClosed   CircuitState = "closed"
	CircuitDraining CircuitState = "draining"
	CircuitIsolated CircuitState = "isolated"
	CircuitOpen     CircuitState = "open"
)

type State struct {
	mu       sync.RWMutex
	circuits map[string]CircuitState
}

func NewState(colors ...string) *State {
	state := &State{circuits: map[string]CircuitState{}}
	for _, color := range colors {
		state.circuits[color] = CircuitClosed
	}
	return state
}

func (state *State) Set(color string, value CircuitState) {
	state.mu.Lock()
	defer state.mu.Unlock()
	state.circuits[color] = value
}

func (state *State) Get(color string) CircuitState {
	state.mu.RLock()
	defer state.mu.RUnlock()
	return state.circuits[color]
}

func (state *State) Colors() []string {
	state.mu.RLock()
	defer state.mu.RUnlock()
	values := make([]string, 0, len(state.circuits))
	for color := range state.circuits {
		values = append(values, color)
	}
	sort.Strings(values)
	return values
}
