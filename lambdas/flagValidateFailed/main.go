package main

import (
	"context"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

var TABLE_NAME = "validationStatus"

func dflCtx() context.Context {
	return context.TODO()
}

type ValidationFailFlagRequest struct {
	TransactionUuid int64 `json:"transactionUuid"`
	Reason          int32 `json:"reason"`
	Error           struct {
		Error string `json:"Error"`
	} `json:"error,omitempty"`
}

func getKey(uuid int64) (map[string]types.AttributeValue, error) {
	sru, err := attributevalue.Marshal(uuid)
	return map[string]types.AttributeValue{"StoreRequestUUID": sru}, err
}

func updateTuple(dyndb *dynamodb.Client, uuid int64, reason int32) error {
	key, err := getKey(uuid)
	if err != nil {
		return err
	}

	update := expression.Set(expression.Name("StatusReason"), expression.Value(reason))
	expr, err := expression.NewBuilder().WithUpdate(update).Build()
	if err != nil {
		return err
	}

	_, err = dyndb.UpdateItem(dflCtx(), &dynamodb.UpdateItemInput{
		TableName:                 &TABLE_NAME,
		Key:                       key,
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		UpdateExpression:          expr.Update(),
	})

	return err
}

func newDynamoDbService() (*dynamodb.Client, error) {
	awsConfig, err := config.LoadDefaultConfig(
		dflCtx(),
		config.WithRegion(os.Getenv("AWS_REGION")))
	if err != nil {
		return nil, err
	}

	return dynamodb.NewFromConfig(awsConfig), nil
}

func handler(e ValidationFailFlagRequest) (bool, error) {
	ddbSvc, err := newDynamoDbService()
	if err != nil {
		return false, err
	}

	err = updateTuple(ddbSvc, e.TransactionUuid, e.Reason)
	return err == nil, err
}

func main() {
	lambda.Start(handler)
}
