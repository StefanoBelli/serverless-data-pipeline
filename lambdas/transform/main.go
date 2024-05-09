package main

import (
	"github.com/aws/aws-lambda-go/lambda"
)

type TupleTransformationRequest struct {
	Success       bool   `json:"success"`
	ValidationIdx int    `json:"validationIdx"`
	Tuple         string `json:"tuple"`
}

type TupleTransformationResponse struct {
	Success           bool   `json:"success"`
	ValidationIdx     int    `json:"validationIdx"`
	TransformationIdx int    `json:"transformationIdx"`
	Tuple             string `json:"tuple"`
}

func handler(e TupleTransformationRequest) (TupleTransformationResponse, error) {
	return TupleTransformationResponse{Success: true, ValidationIdx: 1, TransformationIdx: 1, Tuple: ""}, nil
}

func main() {
	lambda.Start(handler)
}
