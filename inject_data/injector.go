package main

import "fmt"

func inject(path string) error {
	var genChans GeneratorChannels
	genChans.outEntry = make(chan string)
	genChans.outErr = make(chan error)

	go generate(path, []NoiseGenerator{}, genChans)

	for entry := range genChans.outEntry {
		fmt.Println(entry)
	}

	myErr := <-genChans.outErr
	close(genChans.outErr)

	return myErr
}
