package app

import (
	"math/rand"
)

func PickRandom[T any](alternatives []*T, weightSelector func(*T) float64) *T {
	if len(alternatives) == 0 {
		panic("Cannot pick random from empty slice")
	}

	totalWeight := 0.0
	for _, alternative := range alternatives {
		totalWeight += weightSelector(alternative)
	}

	r := rand.Float64() * totalWeight
	cumulativeWeight := 0.0

	for _, alternative := range alternatives {
		cumulativeWeight += weightSelector(alternative)
		if r <= cumulativeWeight {
			return alternative
		}
	}

	return alternatives[len(alternatives)-1]
}
