#!/bin/sh

SOURCES="main.go"

OUTPUT=bin

echo "+++ output directory set to $OUTPUT"

mkdir $OUTPUT

echo " - building $OUTPUT/inspect"

cd src

go build -o ../$OUTPUT/inspect $SOURCES

cd ..
