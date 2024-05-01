package main

import (
	"errors"
	"math/rand"
	"time"
)

type NoiseGenerator struct {
	columnName string
	callback   func(string) string
}

func findNoiseGenerator(gs []NoiseGenerator, c CsvEntry) (NoiseGenerator, error) {
	for _, g := range gs {
		if g.columnName == c.columnName {
			return g, nil
		}
	}

	return NoiseGenerator{"", nil}, errors.New("no matching NoiseGenerator")
}

type GeneratorChannels struct {
	outEntry chan string
	outErr   chan error
}

func generate(filePath string, noiseGens []NoiseGenerator, chans GeneratorChannels) {
	needsDirtyData := programConfig.generator.dirtyData
	threshDirtyData := programConfig.generator.dirtyThresh
	delay := time.Duration(programConfig.generator.everyMs) * time.Millisecond

	csvCommaCh := programConfig.csv.separator

	csv, err := openCsv(filePath, csvCommaCh)
	if err != nil {
		close(chans.outEntry)
		chans.outErr <- err
		return
	}

	defer csv.close()

	var parseErr error

	for {
		ents, parseErr := csv.readNextLine()
		if parseErr != nil {
			break
		}

		genDirty := needsDirtyData && rand.Float32() > threshDirtyData

		out := ""

		for _, ent := range ents {
			if genDirty {
				if rand.Float32() >= 0.7 {
					gen, genErr := findNoiseGenerator(noiseGens, ent)
					if genErr == nil && gen.callback != nil {
						ent.value = gen.callback(ent.value)
					}
				}
			}
			out += ent.value + csvCommaCh
		}

		chans.outEntry <- out[:len(out)-1]

		time.Sleep(delay)
	}

	close(chans.outEntry)

	var csvEofError *CsvEofError
	if errors.As(parseErr, &csvEofError) {
		chans.outErr <- nil
	} else {
		chans.outErr <- parseErr
	}
}
