package unwind

import (
	"fmt"
	"sync"
	"time"
)

type Roll struct {
	ID        string    `json:"id"`
	WidthMM   float64   `json:"width_mm"`
	Diameter  float64   `json:"diameter_mm"`
	LoadedAt  time.Time `json:"loaded_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

type Loader struct {
	mu    sync.RWMutex
	state *State
	roll  Roll
}

func NewLoader(state *State) *Loader {
	return &Loader{state: state}
}

func (loader *Loader) Load(roll Roll) (uint64, error) {
	if roll.ID == "" || roll.WidthMM <= 0 || roll.Diameter <= 0 {
		return 0, fmt.Errorf("roll dimensions and identity are required")
	}
	if !roll.ExpiresAt.IsZero() && !roll.ExpiresAt.After(roll.LoadedAt) {
		return 0, fmt.Errorf("roll expiration must follow load time")
	}
	generation, err := loader.state.LoadRoll(roll.ID)
	if err != nil {
		return 0, err
	}
	loader.mu.Lock()
	loader.roll = roll
	loader.mu.Unlock()
	return generation, nil
}

func (loader *Loader) Current() (Roll, bool) {
	loader.mu.RLock()
	defer loader.mu.RUnlock()
	return loader.roll, loader.roll.ID != ""
}
