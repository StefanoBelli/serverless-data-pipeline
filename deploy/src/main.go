package main

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

var awsConfig aws.Config

func createDynamoDbs() {
	dynamoDbSvc := dynamodb.NewFromConfig(awsConfig)

	for _, dbs := range dynamoDbTables {
		opOut, err := dynamoDbSvc.CreateTable(context.TODO(), &dbs)
		if err != nil {
			log.Fatalf("unable to create table: %v", err)
		}

		log.Println(
			"table %s status %s",
			*(opOut.TableDescription.TableName),
			opOut.TableDescription.TableStatus)
	}
}

func main() {
	awsCfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion(AWS_REGION))
	if err != nil {
		log.Fatalf("unable to load SDK config: %v", err)
	}

	awsConfig = awsCfg

	createDynamoDbs()
}
