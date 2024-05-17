#!/bin/sh

SOURCES="main.go config.go"

OUTPUT=bin

echo "+++ output directory set to $OUTPUT"

mkdir $OUTPUT

echo " - building $OUTPUT/deploy"

cd src

go build -o ../$OUTPUT/deploy $SOURCES

cd ..
