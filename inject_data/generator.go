package main

type GeneratorReplacer struct {
	columnName string
	callback   func(string) string
}

func generate(outEntry chan (string), r []GeneratorReplacer) {
}
