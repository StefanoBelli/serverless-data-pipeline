package dyndbutils

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

func dflCtx() context.Context {
	return context.TODO()
}

type TupleStatus struct {
	StoreRequestUuid int64  `dynamodbav:"StoreRequestUUID"`
	RawTuple         string `dynamodbav:"RawTuple"`
	StatusReason     int32  `dynamodbav:"StatusReason"`
}

func BuildDefaultTupleStatus(uuid int64, rawTuple *string) TupleStatus {
	return TupleStatus{
		StoreRequestUuid: uuid,
		RawTuple:         *rawTuple,
		StatusReason:     0,
	}
}

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

func NewDynamoDbService() (*dynamodb.Client, error) {
	awsConfig, err := config.LoadDefaultConfig(
		dflCtx(),
		config.WithRegion(os.Getenv("AWS_REGION")))
	if err != nil {
		return nil, err
	}

	return dynamodb.NewFromConfig(awsConfig), nil
}
