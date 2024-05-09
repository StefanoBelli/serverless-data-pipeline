package main

import (
	"github.com/aws/aws-lambda-go/lambda"
)

type StoreFailFlagRequest struct {
	StoreLogIdx int `json:"storeLogIdx"`
}

func handler(e StoreFailFlagRequest) (bool, error) {
	return true, nil
}

func main() {
	lambda.Start(handler)
}
