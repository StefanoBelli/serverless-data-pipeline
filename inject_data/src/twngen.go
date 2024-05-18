package main

import "math/rand"

type TupleWiseNoiseGenerator func(*string)

func generateTupleWiseNoise(t *string, g *[]TupleWiseNoiseGenerator) {
	if programConfig.generator.dirtyData {
		for _, twngen := range *g {
			if rand.Intn(100) == 60 {
				twngen(t)
				return
			}
		}
	}
}
