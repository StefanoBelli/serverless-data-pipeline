package main

import (
	fpf "flagPhaseFailed"

	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	fpf.SetTableName("storeStatus")
	lambda.Start(fpf.Handler)
}
