@echo off

set SOURCES=main.go

set OUTPUT=bin

echo +++ output directory set to %OUTPUT%

md %OUTPUT%

echo  - building %OUTPUT%/inspect.exe

cd src

go build -o ../%OUTPUT%/inspect.exe %SOURCES%

cd ..