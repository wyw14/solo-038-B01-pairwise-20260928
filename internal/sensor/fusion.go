package sensor

import (
	"fmt"
	"math"
	"sort"
)

type SourceVote struct {
	Origin string  `json:"origin"`
	Value  float64 `json:"value"`
	Kinds  int     `json:"kinds"`
}

func IndependentVotes(readings []Reading) ([]SourceVote, error) {
	byOrigin := map[string][]Reading{}
	for _, reading := range readings {
		if err := reading.Validate(); err != nil {
			return nil, err
		}
		byOrigin[reading.Origin] = append(byOrigin[reading.Origin], reading)
	}
	votes := make([]SourceVote, 0, len(byOrigin))
	for origin, values := range byOrigin {
		total := 0.0
		kinds := map[ReadingKind]bool{}
		for _, value := range values {
			total += value.Value
			kinds[value.Kind] = true
		}
		votes = append(votes, SourceVote{Origin: origin, Value: total / float64(len(values)), Kinds: len(kinds)})
	}
	sort.Slice(votes, func(left, right int) bool { return votes[left].Origin < votes[right].Origin })
	return votes, nil
}

func Agree(votes []SourceVote, tolerance float64) (bool, error) {
	if len(votes) < 2 {
		return false, nil
	}
	if tolerance < 0 {
		return false, fmt.Errorf("tolerance cannot be negative")
	}
	minimum, maximum := votes[0].Value, votes[0].Value
	for _, vote := range votes[1:] {
		minimum = math.Min(minimum, vote.Value)
		maximum = math.Max(maximum, vote.Value)
	}
	return maximum-minimum <= tolerance, nil
}
