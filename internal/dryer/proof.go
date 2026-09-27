package dryer

import (
	"fmt"
	"time"

	"github.com/wyw14/solo-038/internal/model"
)

type IntervalProof struct {
	Proof       model.Proof `json:"proof"`
	Temperature float64     `json:"temperature"`
	Airflow     float64     `json:"airflow"`
}

func BuildProof(job model.Job, interval model.Interval, temperature, airflow float64, now time.Time) (IntervalProof, error) {
	if err := interval.Validate(); err != nil {
		return IntervalProof{}, err
	}
	if temperature <= 0 || airflow <= 0 {
		return IntervalProof{}, fmt.Errorf("dryer temperature and airflow proof are required")
	}
	proof := model.Proof{Kind: "dryer", Source: "dryer-zones", JobGeneration: job.Generation, RollGeneration: job.RollGeneration, ValidFrom: interval.StartMeter, ValidTo: interval.EndMeter, ObservedAt: now.UTC()}
	return IntervalProof{Proof: proof, Temperature: temperature, Airflow: airflow}, nil
}

func (proof IntervalProof) Covers(interval model.Interval) bool {
	return proof.Proof.JobGeneration == interval.JobGeneration && proof.Proof.RollGeneration == interval.RollGeneration && proof.Proof.ValidFrom <= interval.StartMeter && proof.Proof.ValidTo >= interval.EndMeter
}
