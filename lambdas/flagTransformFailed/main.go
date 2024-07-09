package main

import (
	fpf "flagPhaseFailed"

	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	// Refer to local package flagPhaseFailed (located at ../flagPhaseFailed)
	fpf.SetTableName("transformationStatus")
	lambda.Start(fpf.Handler)
}
