package main

import (
	"github.com/aws/aws-lambda-go/lambda"
)

type TupleValidationRequest struct {
	Tuple string `json:"tuple"`
}

type TupleValidationResponse struct {
	Success       bool   `json:"success"`
	ValidationIdx int    `json:"validationIdx"`
	Tuple         string `json:"tuple"`
}

func handler(e TupleValidationRequest) (TupleValidationResponse, error) {
	return TupleValidationResponse{Success: true, ValidationIdx: 1, Tuple: ""}, nil
}

func main() {
	lambda.Start(handler)
}
