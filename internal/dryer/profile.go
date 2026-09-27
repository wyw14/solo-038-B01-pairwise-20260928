package dryer

import (
	"fmt"
	"sort"
)

type Zone struct {
	Name        string  `json:"name"`
	Temperature float64 `json:"temperature"`
	Airflow     float64 `json:"airflow"`
}

type Profile struct {
	Name  string `json:"name"`
	Zones []Zone `json:"zones"`
}

func (profile Profile) Validate() error {
	if profile.Name == "" || len(profile.Zones) == 0 {
		return fmt.Errorf("dryer profile name and zones are required")
	}
	seen := map[string]bool{}
	for _, zone := range profile.Zones {
		if zone.Name == "" || zone.Temperature < 0 || zone.Airflow < 0 || seen[zone.Name] {
			return fmt.Errorf("dryer profile zone is invalid")
		}
		seen[zone.Name] = true
	}
	return nil
}

func (profile Profile) MaximumTemperature() float64 {
	values := make([]float64, 0, len(profile.Zones))
	for _, zone := range profile.Zones {
		values = append(values, zone.Temperature)
	}
	sort.Float64s(values)
	if len(values) == 0 {
		return 0
	}
	return values[len(values)-1]
}

func (profile Profile) AverageAirflow() float64 {
	if len(profile.Zones) == 0 {
		return 0
	}
	total := 0.0
	for _, zone := range profile.Zones {
		total += zone.Airflow
	}
	return total / float64(len(profile.Zones))
}
