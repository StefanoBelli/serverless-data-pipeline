package main

import (
	fpf "flagPhaseFailed"

	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	fpf.SetTableName("transformationStatus")
	lambda.Start(fpf.Handler)
}
