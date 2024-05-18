package main

import (
	"errors"
	"math/rand"
	"time"
)

type ColumnNoiseGenerator struct {
	columnName string
	callback   func(string) string
}

type ColumnNoiseGenerationChannels struct {
	outEntry chan string
	outErr   chan error
}

func readTuplesAndGenerateColumnNoise(
	filePath string, noiseGens []ColumnNoiseGenerator, chans ColumnNoiseGenerationChannels) {

	delay := time.Duration(programConfig.generator.everyMs) * time.Millisecond

	csv, err := openCsv(filePath, programConfig.csv.separator)
	if err != nil {
		closeChansWithErr(err, &chans)
		return
	}

	defer csv.close()

	err = discardFirstEntriesAsRequired(&csv)
	if err != nil {
		closeChansWithErr(err, &chans)
		return
	}

	var parseErr error

	for {
		ents, parseErr := csv.readNextLine()
		if parseErr != nil {
			break
		}

		out := generateColumnNoise(&ents, &noiseGens)
		chans.outEntry <- out[:len(out)-1]
		time.Sleep(delay)
	}

	closeChansWithErr(parseErr, &chans)
}

func closeChansWithErr(myerr error, chs *ColumnNoiseGenerationChannels) {
	close(chs.outEntry)

	var csvEofError *CsvEofError
	if errors.As(myerr, &csvEofError) {
		chs.outErr <- nil
	} else {
		chs.outErr <- myerr
	}
}

func discardFirstEntriesAsRequired(csv *Csv) error {
	for range programConfig.injector.startAt {
		_, err := csv.readNextLine()
		if err != nil {
			return err
		}
	}

	return nil
}

func generateColumnNoise(ents *[]CsvEntry, noiseGens *[]ColumnNoiseGenerator) string {
	needsDirtyData := programConfig.generator.dirtyData
	threshDirtyData := programConfig.generator.dirtyThresh

	genDirty := needsDirtyData && rand.Float32() > threshDirtyData

	out := ""

	for _, ent := range *ents {
		if genDirty {
			if rand.Float32() >= 0.7 {
				gen, genErr := findColumnNoiseGenerator(noiseGens, &ent)
				if genErr == nil && gen.callback != nil {
					ent.value = gen.callback(ent.value)
				}
			}
		}
		out += ent.value + programConfig.csv.separator
	}

	return out
}

func findColumnNoiseGenerator(gs *[]ColumnNoiseGenerator, c *CsvEntry) (ColumnNoiseGenerator, error) {
	for _, g := range *gs {
		if g.columnName == c.columnName {
			return g, nil
		}
	}

	return ColumnNoiseGenerator{"", nil}, errors.New("no matching NoiseGenerator")
}
