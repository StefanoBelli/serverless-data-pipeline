package main

import (
	"github.com/aws/aws-lambda-go/lambda"
)

type TransformationFailFlagRequest struct {
	TransformationIdx int `json:"transformationIdx"`
}

func handler(e TransformationFailFlagRequest) (bool, error) {
	return true, nil
}

func main() {
	lambda.Start(handler)
}
