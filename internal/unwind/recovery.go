package unwind

import "fmt"

type RecoveryRecord struct {
	Roll           string  `json:"roll"`
	RollGeneration uint64  `json:"roll_generation"`
	Owner          string  `json:"owner"`
	Speed          float64 `json:"speed"`
}

func Restore(state *State, record RecoveryRecord) error {
	if record.Roll == "" || record.RollGeneration == 0 || record.Owner == "" {
		return fmt.Errorf("recovery record is incomplete")
	}
	state.mu.Lock()
	defer state.mu.Unlock()
	state.activeRoll = record.Roll
	state.rollGeneration = record.RollGeneration
	state.feedOwner = record.Owner
	state.feedSpeed = record.Speed
	return nil
}

func ResumeAtSafeSpeed(state *State, owner string, generation uint64, requested, maximum float64) error {
	if maximum <= 0 {
		return fmt.Errorf("maximum recovery speed is required")
	}
	if requested > maximum {
		requested = maximum
	}
	return state.StartFeed(owner, generation, requested)
}
