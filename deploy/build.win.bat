@echo off

set SOURCES=src/main.go

set OUTPUT=bin

echo +++ output directory set to %OUTPUT%

md %OUTPUT%

echo  - building %OUTPUT%/deploy.exe

go build -o %OUTPUT%/deploy.exe %SOURCES%