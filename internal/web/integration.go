package web

import (
	"context"
	"fmt"
	"time"
)

type WorkerResult struct {
	Samples int     `json:"samples"`
	Last    float64 `json:"last"`
	Stopped bool    `json:"stopped"`
}

func RunControlLoop(ctx context.Context, interval time.Duration, target float64, apply func(float64)) <-chan WorkerResult {
	result := make(chan WorkerResult, 1)
	go func() {
		defer close(result)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		current := 0.0
		samples := 0
		for {
			select {
			case <-ctx.Done():
				result <- WorkerResult{Samples: samples, Last: current, Stopped: true}
				return
			case <-ticker.C:
				current += (target - current) * 0.25
				apply(current)
				samples++
			}
		}
	}()
	return result
}

func AwaitWorker(ctx context.Context, result <-chan WorkerResult) (WorkerResult, error) {
	select {
	case value, ok := <-result:
		if !ok {
			return WorkerResult{}, fmt.Errorf("worker returned no result")
		}
		return value, nil
	case <-ctx.Done():
		return WorkerResult{}, ctx.Err()
	}
}
