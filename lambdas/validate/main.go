package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const CSV_COMMA_SEP = ","
const CSV_NUMCOLS = 20

var TABLE_NAME = "validationStatus"

func dflCtx() context.Context {
	return context.TODO()
}

type TupleValidationRequest struct {
	Tuple string `json:"tuple"`
}

type TupleValidationResponse struct {
	Success         bool   `json:"success"`
	TransactionUuid int64  `json:"transactionUuid"`
	Tuple           string `json:"tuple"`
}

func validResponse(uuid int64, rawTuple *string) (TupleValidationResponse, error) {
	return TupleValidationResponse{
		Success:         true,
		TransactionUuid: uuid,
		Tuple:           *rawTuple,
	}, nil
}

func invalidResponse(uuid int64) (TupleValidationResponse, error) {
	return TupleValidationResponse{
		Success:         false,
		TransactionUuid: uuid,
	}, nil
}

func erroredResponse(msg string, err error) (TupleValidationResponse, error) {
	return TupleValidationResponse{}, fmt.Errorf("%s: %v", msg, err)
}

func calculateTransactionUuid(rawTuple *string, beginTime int64) int64 {
	var uuid int64
	for i, c := range *rawTuple {
		uuid += int64(c) * int64(i)
	}

	return uuid
}

func putRawTuple(dyndb *dynamodb.Client, uuid int64, rawTuple *string) error {
	_, err := dyndb.PutItem(dflCtx(), &dynamodb.PutItemInput{
		Item: map[string]types.AttributeValue{
			strconv.FormatInt(uuid, 10): &types.AttributeValueMemberN{},
			*rawTuple:                   &types.AttributeValueMemberS{},
			"0":                         &types.AttributeValueMemberN{},
		},
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

func fieldChecksAreOk(tuple *string) bool {
	csvCols := strings.Split(*tuple, CSV_COMMA_SEP)
	if len(csvCols) != CSV_NUMCOLS {
		return false
	}

	for _, column := range singleColumnCheckers {
		if !column.check(&csvCols[column.idx]) {
			return false
		}
	}

	for _, columns := range crossColumnCheckers {
		var kols []*string
		for _, j := range columns.idxs {
			kols = append(kols, &csvCols[j])
		}

		if !columns.check(&kols) {
			return false
		}
	}

	return true
}

func handler(e TupleValidationRequest) (TupleValidationResponse, error) {
	transactionBeginTime := time.Now().UnixNano()

	fixedTuple := strings.TrimSpace(e.Tuple)
	if len(fixedTuple) == 0 {
		return erroredResponse("receiving input",
			errors.New("empty tuple"))
	}

	ddbSvc, err := newDynamoDbService()
	if err != nil {
		return erroredResponse("unable to load dynamodb service", err)
	}

	transactionUuid := calculateTransactionUuid(&e.Tuple, transactionBeginTime)

	err = putRawTuple(ddbSvc, transactionUuid, &fixedTuple)
	if err != nil {
		return erroredResponse("unable to put raw tuple", err)
	}

	if fieldChecksAreOk(&fixedTuple) {
		return validResponse(transactionUuid, &fixedTuple)
	} else {
		return invalidResponse(transactionUuid)
	}
}

func main() {
	lambda.Start(handler)
}
