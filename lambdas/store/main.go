package main

import (
	"github.com/aws/aws-lambda-go/lambda"
)

type TupleStoreRequest struct {
	Success           bool   `json:"success"`
	ValidationIdx     int    `json:"validationIdx"`
	TransformationIdx int    `json:"transformationIdx"`
	Tuple             string `json:"tuple"`
}

type TupleStoreResponse struct {
	Success           bool `json:"success"`
	ValidationIdx     int  `json:"validationIdx"`
	TransformationIdx int  `json:"transformationIdx"`
	StoreLogIdx       int  `json:"storeLogIdx"`
}

func handler(e TupleStoreRequest) (TupleStoreResponse, error) {
	return TupleStoreResponse{Success: true, ValidationIdx: 1, TransformationIdx: 1, StoreLogIdx: 1}, nil
}

func main() {
	lambda.Start(handler)
}
