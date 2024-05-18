#!/bin/bash

LAMBDAS="
    validate 
    transform
    store
    flagValidateFailed
    flagTransformFailed
    flagStoreFailed
    authorizer
"

OUTPUT=pkgs

FAILSIMFLAG=,ENABLE_FAILSIM

mkdir $OUTPUT

echo "+++ output directory set to $OUTPUT"

export GOOS=linux
export GOARCH=amd64
export CGO_ENABLED=0

build() {
    echo failsimflag=$FAILSIMFLAG

    for l in $@; do
        echo building lambda $l...

        cd $l

        BOOTSTRAP=../$OUTPUT/$l/bin-$GOOS-$GOARCH/bootstrap
        ZIP=../$OUTPUT/$l/$l.zip
        SOURCE=main.go

        echo " - building"
        go build -tags=lambda.norpc$FAILSIMFLAG -o $BOOTSTRAP $SOURCE

        echo " - packaging"
        zip -r -j $ZIP $BOOTSTRAP

        echo " + package $OUTPUT/$l/$l.zip ready to upload"

        cd ..
    done
}

if [ $# -ge 1 ]; then
    for lambda in $LAMBDAS; do
        if [[ $1 == $lambda ]]; then
            if [[ $2 == "-d" ]]; then
                FAILSIMFLAG=""
            fi

            build $1
            exit 0
        fi
    done

    if [[ $1 == "-d" ]]; then
        FAILSIMFLAG=""
        build $LAMBDAS
        exit 0
    fi 

    echo $1 unknown lambda
    echo exiting now...
    exit 1
fi

build $LAMBDAS
