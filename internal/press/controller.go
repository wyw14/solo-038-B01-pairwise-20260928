package press

import (
	"fmt"
	"sort"
	"sync"
)

type Unit struct {
	Name       string  `json:"name"`
	Speed      float64 `json:"speed"`
	Enabled    bool    `json:"enabled"`
	ZeroSensor bool    `json:"zero_sensor"`
}

type Controller struct {
	mu    sync.RWMutex
	units map[string]Unit
}

func NewController(names ...string) (*Controller, error) {
	if len(names) == 0 {
		return nil, fmt.Errorf("at least one press unit is required")
	}
	controller := &Controller{units: map[string]Unit{}}
	for _, name := range names {
		if name == "" {
			return nil, fmt.Errorf("press unit name is required")
		}
		controller.units[name] = Unit{Name: name, Enabled: true}
	}
	return controller, nil
}

func (controller *Controller) SetSpeed(name string, speed float64) error {
	controller.mu.Lock()
	defer controller.mu.Unlock()
	unit, ok := controller.units[name]
	if !ok || speed < 0 {
		return fmt.Errorf("press speed update is invalid")
	}
	unit.Speed = speed
	unit.ZeroSensor = speed == 0
	controller.units[name] = unit
	return nil
}

func (controller *Controller) Unit(name string) (Unit, bool) {
	controller.mu.RLock()
	defer controller.mu.RUnlock()
	unit, ok := controller.units[name]
	return unit, ok
}

func (controller *Controller) Names() []string {
	controller.mu.RLock()
	defer controller.mu.RUnlock()
	names := make([]string, 0, len(controller.units))
	for name := range controller.units {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
