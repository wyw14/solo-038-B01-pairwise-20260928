package ink

import (
	"fmt"
	"sync"
)

type Circuit struct {
	mu        sync.Mutex
	state     *State
	drainFail map[string]error
}

func NewCircuit(state *State) *Circuit {
	return &Circuit{state: state, drainFail: map[string]error{}}
}

func (circuit *Circuit) FailDrain(color string, err error) {
	circuit.mu.Lock()
	defer circuit.mu.Unlock()
	circuit.drainFail[color] = err
}

func (circuit *Circuit) Drain(color string) error {
	circuit.mu.Lock()
	defer circuit.mu.Unlock()
	if circuit.state.Get(color) != CircuitOpen {
		return fmt.Errorf("ink circuit %s is not open", color)
	}
	circuit.state.Set(color, CircuitDraining)
	if err := circuit.drainFail[color]; err != nil {
		delete(circuit.drainFail, color)
		circuit.state.Set(color, CircuitOpen)
		return err
	}
	circuit.state.Set(color, CircuitIsolated)
	return nil
}

func (circuit *Circuit) Open(color string) error {
	current := circuit.state.Get(color)
	if current != CircuitClosed && current != CircuitIsolated {
		return fmt.Errorf("ink circuit %s cannot open from %s", color, current)
	}
	circuit.state.Set(color, CircuitOpen)
	return nil
}
