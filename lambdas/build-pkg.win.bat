@echo off

setlocal enabledelayedexpansion

set lambdas[0]=validate
set lambdas[1]=transform
set lambdas[2]=store
set lambdas[3]=flagValidateFailed
set lambdas[4]=flagTransformFailed
set lambdas[5]=flagStoreFailed

set OUTPUT=pkgs

md %OUTPUT%

echo +++ output directory set to %OUTPUT%

set GOOS=linux
set GOARCH=amd64
set CGO_ENABLED=0

set start=0
set end=5

if "%~1"=="" goto all

for /l %%i in (0,1,5) do (
    if !lambdas[%%i]! == %1 (
        Set start=%%i
        Set end=%%i
    )
)

if !start! NEQ !end! (
    echo %1 unknown lambda
    echo exiting now...
    exit 1
)

:all

for /l %%i in (!start!,1,!end!) do (
    echo building lambda !lambdas[%%i]!...

    cd !lambdas[%%i]!

    set BOOTSTRAP=../%OUTPUT%/!lambdas[%%i]!/bin-%GOOS%-%GOARCH%/bootstrap 
    set ZIP=../%OUTPUT%/!lambdas[%%i]!/!lambdas[%%i]!.zip
    set SOURCE=main.go

    echo  - building
    go build -tags lambda.norpc -o !BOOTSTRAP! !SOURCE!

    echo  - packaging
    tar.exe -a -c -f !ZIP! !BOOTSTRAP!

    echo  + package %OUTPUT%/!lambdas[%%i]!/!lambdas[%%i]!.zip ready to upload

    cd ..
)