@echo off

set SOURCES=src/csv_parser.go src/generator.go src/injector.go src/main.go

set OUTPUT=bin

echo +++ output directory set to %OUTPUT%

md %OUTPUT%

echo  - building %OUTPUT%/inject_data.exe

go build -o %OUTPUT%/inject_data.exe %SOURCES%