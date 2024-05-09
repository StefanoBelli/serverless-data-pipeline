package main

import (
	"github.com/aws/aws-lambda-go/lambda"
)

type ValidationFailFlagRequest struct {
	ValidationIdx int `json:"validationIdx"`
}

func handler(e ValidationFailFlagRequest) (bool, error) {
	return true, nil
}

func main() {
	lambda.Start(handler)
}
