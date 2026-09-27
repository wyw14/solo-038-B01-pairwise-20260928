package unwind

import (
	"fmt"
	"sort"
	"sync"
)

type Reservation struct {
	Operation string  `json:"operation"`
	Meters    float64 `json:"meters"`
	Fence     uint64  `json:"fence"`
}

type Festoon struct {
	mu           sync.Mutex
	capacity     float64
	fence        uint64
	reservations map[string]Reservation
}

func NewFestoon(capacity float64) *Festoon {
	return &Festoon{capacity: capacity, reservations: map[string]Reservation{}}
}

func (festoon *Festoon) Reserve(operation string, meters float64) (Reservation, error) {
	festoon.mu.Lock()
	defer festoon.mu.Unlock()
	if operation == "" || meters <= 0 {
		return Reservation{}, fmt.Errorf("reservation operation and meters are required")
	}
	if existing, ok := festoon.reservations[operation]; ok {
		return existing, nil
	}
	if meters > festoon.availableLocked() {
		return Reservation{}, fmt.Errorf("festoon capacity exceeded")
	}
	festoon.fence++
	reservation := Reservation{Operation: operation, Meters: meters, Fence: festoon.fence}
	festoon.reservations[operation] = reservation
	return reservation, nil
}

func (festoon *Festoon) Release(operation string, fence uint64) bool {
	festoon.mu.Lock()
	defer festoon.mu.Unlock()
	reservation, ok := festoon.reservations[operation]
	if !ok || reservation.Fence != fence {
		return false
	}
	delete(festoon.reservations, operation)
	return true
}

func (festoon *Festoon) Available() float64 {
	festoon.mu.Lock()
	defer festoon.mu.Unlock()
	return festoon.availableLocked()
}

func (festoon *Festoon) availableLocked() float64 {
	used := 0.0
	for _, reservation := range festoon.reservations {
		used += reservation.Meters
	}
	return festoon.capacity - used
}

func (festoon *Festoon) Reservations() []Reservation {
	festoon.mu.Lock()
	defer festoon.mu.Unlock()
	values := make([]Reservation, 0, len(festoon.reservations))
	for _, reservation := range festoon.reservations {
		values = append(values, reservation)
	}
	sort.Slice(values, func(left, right int) bool { return values[left].Fence < values[right].Fence })
	return values
}
