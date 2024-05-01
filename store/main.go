package main

import (
	"github.com/aws/aws-lambda-go/lambda"
)

type MyEvent struct {
	Content string `json:"content"`
}

func handler(e MyEvent) (string, error) {
	return e.Content, nil
}

func main() {
	lambda.Start(handler)
}
