@echo off

set SOURCES=main.go config.go cstrs.go

set OUTPUT=bin

echo +++ output directory set to %OUTPUT%

md %OUTPUT%

echo  - building %OUTPUT%/deploy.exe

cd src

go build -o ../%OUTPUT%/deploy.exe %SOURCES%

cd ..