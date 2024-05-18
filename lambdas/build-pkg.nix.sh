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

mkdir $OUTPUT

echo "+++ output directory set to $OUTPUT"

export GOOS=linux
export GOARCH=amd64
export CGO_ENABLED=0

build() {
    for l in $@; do
        echo building lambda $l...

        cd $l

        BOOTSTRAP=../$OUTPUT/$l/bin-$GOOS-$GOARCH/bootstrap
        ZIP=../$OUTPUT/$l/$l.zip
        SOURCE=main.go

        echo " - building"
        go build -tags=lambda.norpc,ENABLE_FAILSIM -o $BOOTSTRAP $SOURCE

        echo " - packaging"
        zip -r -j $ZIP $BOOTSTRAP

        echo " + package $OUTPUT/$l/$l.zip ready to upload"

        cd ..
    done
}

if [ $# -ge 1 ]; then
    for lambda in $LAMBDAS; do
        if [[ $1 == $lambda ]]; then
            build $1
            exit 0
        fi
    done

    echo $1 unknown lambda
    echo exiting now...
    exit 1
fi

build $LAMBDAS
