/*
 * package contains exported and commonly-used functions to avoid excessive
 * code duplications. Code is self-explainatory and easy to read
 */
package dyndbutils

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

/* not exported */
func dflCtx() context.Context {
	return context.TODO()
}

type tupleStatus struct {
	StoreRequestId uint64 `dynamodbav:"StoreRequestId"`
	RawTuple       string `dynamodbav:"RawTuple"`
	StatusReason   int32  `dynamodbav:"StatusReason"`
}

/* exported */

// Build a tuple with no error (transaction status: success)
func BuildDefaultTupleStatus(id uint64, rawTuple *string) interface{} {
	return tupleStatus{
		StoreRequestId: id,
		RawTuple:       *rawTuple,
		StatusReason:   0,
	}
}

// Put an item in whatever dynamodb table
func PutInTable(dyndb *dynamodb.Client, ent interface{}, table *string) error {
	item, err := attributevalue.MarshalMap(ent)
	if err != nil {
		return err
	}

	_, err = dyndb.PutItem(dflCtx(), &dynamodb.PutItemInput{
		Item:      item,
		TableName: table,
	})

	return err
}

// Obtain a new DynamoDB client
func NewDynamoDbService() (*dynamodb.Client, error) {
	// the "light" VM which runs this lambda has AWS_REGION env var set
	awsConfig, err := config.LoadDefaultConfig(
		dflCtx(),
		config.WithRegion(os.Getenv("AWS_REGION")))
	if err != nil {
		return nil, err
	}

	return dynamodb.NewFromConfig(awsConfig), nil
}
