#!/bin/sh

SOURCES="csv_parser.go cngen.go twngen.go injector.go main.go"

OUTPUT=bin

echo "+++ output directory set to $OUTPUT"

mkdir $OUTPUT

echo " - building $OUTPUT/inject_data"

cd src

go build -o ../$OUTPUT/inject_data $SOURCES

cd ..
