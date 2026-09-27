package sensor

import (
	"sort"
	"sync"
	"time"
)

type State struct {
	mu       sync.RWMutex
	readings map[string]Reading
	frames   map[string]CameraFrame
}

func NewState() *State {
	return &State{readings: map[string]Reading{}, frames: map[string]CameraFrame{}}
}

func (state *State) PutReading(reading Reading) {
	state.mu.Lock()
	defer state.mu.Unlock()
	state.readings[reading.ID] = reading
}

func (state *State) PutFrame(frame CameraFrame) {
	state.mu.Lock()
	defer state.mu.Unlock()
	state.frames[frame.ID] = frame
}

func (state *State) CurrentReadings(now time.Time, maximumAge time.Duration) []Reading {
	state.mu.RLock()
	defer state.mu.RUnlock()
	values := []Reading{}
	for _, reading := range state.readings {
		if now.Sub(reading.ObservedAt) <= maximumAge {
			values = append(values, reading)
		}
	}
	sort.Slice(values, func(left, right int) bool { return values[left].ID < values[right].ID })
	return values
}

func (state *State) FrameCount() int {
	state.mu.RLock()
	defer state.mu.RUnlock()
	return len(state.frames)
}
