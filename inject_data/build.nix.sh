#!/bin/sh

SOURCES="src/csv_parser.go src/generator.go src/injector.go src/main.go"

OUTPUT=bin

echo "+++ output directory set to $OUTPUT"

mkdir $OUTPUT

echo " - building $OUTPUT/inject_data"

go build -o $OUTPUT/inject_data $SOURCES