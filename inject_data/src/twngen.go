package main

import "math/rand/v2"

type TupleWiseNoiseGenerator func(*string)

func generateTupleWiseNoise(t *string, g *[]TupleWiseNoiseGenerator) {
	if programConfig.generator.dirtyData {
		for _, twngen := range *g {
			if rand.IntN(100) == 60 {
				twngen(t)
				return
			}
		}
	}
}
