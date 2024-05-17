package flagPhaseFailed

import (
	"context"
	"errors"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

/* not exported */

var tableName = ""
var tableNameIsSet = false

func dflCtx() context.Context {
	return context.TODO()
}

type failFlagRequest struct {
	TransactionId uint64 `json:"transactionId"`
	Reason        int32  `json:"reason"`
	Error         struct {
		Error string `json:"Error"`
	} `json:"error,omitempty"`
}

func getKey(id uint64) (map[string]types.AttributeValue, error) {
	sru, err := attributevalue.Marshal(id)
	return map[string]types.AttributeValue{"StoreRequestId": sru}, err
}

func updateTuple(dyndb *dynamodb.Client, id uint64, reason int32) error {
	key, err := getKey(id)
	if err != nil {
		return err
	}

	update := expression.Set(expression.Name("StatusReason"), expression.Value(reason))
	expr, err := expression.NewBuilder().WithUpdate(update).Build()
	if err != nil {
		return err
	}

	_, err = dyndb.UpdateItem(dflCtx(), &dynamodb.UpdateItemInput{
		TableName:                 &tableName,
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

func getReasonCodeFromErrorType(errorType *string) int32 {
	switch *errorType {
	case "TransformError":
		return 2
	case "StoreError":
		return 3
	}

	return 4 //Unknown
}

/* exported */

func Handler(e failFlagRequest) (bool, error) {
	if tableNameIsSet {
		if e.Reason == 0 {
			e.Reason = getReasonCodeFromErrorType(&e.Error.Error)
		}

		ddbSvc, err := newDynamoDbService()
		if err != nil {
			return false, err
		}

		err = updateTuple(ddbSvc, e.TransactionId, e.Reason)
		return err == nil, err
	} else {
		return false, errors.New("need to set tableName (dev error)")
	}
}

func SetTableName(yourTableName string) {
	tableName = yourTableName
	tableNameIsSet = true
}
