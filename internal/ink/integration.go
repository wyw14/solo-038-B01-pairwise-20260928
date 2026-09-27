package ink

import "fmt"

func RequireReady(state *State, colors ...string) error {
	for _, color := range colors {
		if state.Get(color) != CircuitIsolated {
			return fmt.Errorf("ink color %s is not isolated", color)
		}
	}
	return nil
}

func ActiveCircuitReady(state *State, color string) bool {
	return state.Get(color) == CircuitOpen
}
