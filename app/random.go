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

func Shuffle[T any](items []T) []T {
	// Iterate from the end to the beginning of the slice
	for i := len(items) - 1; i > 0; i-- {
		// Generate a random index between 0 and i (inclusive)
		j := rand.Intn(i + 1)

		// Swap the items at index i and j
		items[i], items[j] = items[j], items[i]
	}
	return items
}
