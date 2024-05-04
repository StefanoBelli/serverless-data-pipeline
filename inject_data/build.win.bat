@echo off

set SOURCES=csv_parser.go generator.go injector.go main.go

set OUTPUT=bin

echo +++ output directory set to %OUTPUT%

md %OUTPUT%

echo  - building %OUTPUT%/inject_data.exe

cd src

go build -o ../%OUTPUT%/inject_data.exe %SOURCES%

cd ..