#!/bin/sh

SOURCES="src/main.go src/config.go"

OUTPUT=bin

echo "+++ output directory set to $OUTPUT"

mkdir $OUTPUT

echo " - building $OUTPUT/deploy"

go build -o $OUTPUT/deploy $SOURCES