package main

type GeneratorRule struct {
	columnName string
	replacer   func(string) string
}

type GeneratorConfig struct {
	filePath       string
	everyMs        int
	dirtyData      bool
	dirtyThreshold float32
	outEntry       chan (string)
}

func generate(c GeneratorConfig, r []GeneratorRule) {

}
