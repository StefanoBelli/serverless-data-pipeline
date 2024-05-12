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

type SingleColumnChecker struct {
	idx   int
	check func(*string) bool
}

var singleColumnCheckers = []SingleColumnChecker{
	{
		idx: 0,
		check: func(s *string) bool {
			i, err := strconv.ParseInt(*s, 10, 32)
			if err != nil || i < 0 {
				*s = ""
			}

			return true
		},
	},
	{
		idx: 1,
		check: func(s *string) bool {
			i, err := strconv.ParseInt(*s, 10, 32)
			if err != nil || i < 1 || i > 2 {
				*s = ""
			}

			return true
		},
	},
	{
		idx: 4,
		check: func(s *string) bool {
			i, err := strconv.ParseFloat(*s, 32)
			if err != nil {
				*s = ""
			} else if i < 1 || i > 5 {
				return false
			}

			return true
		},
	},
	{
		idx: 5,
		check: func(s *string) bool {
			i, err := strconv.ParseFloat(*s, 32)
			if err != nil || i < 0 {
				return false
			}

			return true
		},
	},
	{
		idx: 6,
		check: func(s *string) bool {
			i, err := strconv.ParseFloat(*s, 32)
			if err != nil || i < 0 || (i > 6 && i != 99) {
				return false
			}

			return true
		},
	},
	{
		idx: 7,
		check: func(s *string) bool {
			if *s != "Y" && *s != "N" {
				*s = ""
			}

			return true
		},
	},
	{
		idx: 8,
		check: func(s *string) bool {
			i, err := strconv.ParseInt(*s, 10, 32)
			if err != nil || i < 0 {
				return false
			}

			return true
		},
	},
	{
		idx: 9,
		check: func(s *string) bool {
			i, err := strconv.ParseInt(*s, 10, 32)
			if err != nil || i < 0 {
				return false
			}

			return true
		},
	},
	{
		idx: 10,
		check: func(s *string) bool {
			i, err := strconv.ParseInt(*s, 10, 32)
			if err != nil || i < 0 || i > 6 {
				*s = ""
			}

			return true
		},
	},
	{
		idx: 11,
		check: func(s *string) bool {
			_, err := strconv.ParseFloat(*s, 32)
			if err != nil {
				*s = ""
			}

			return true
		},
	},
	{
		idx: 12,
		check: func(s *string) bool {
			_, err := strconv.ParseFloat(*s, 32)
			if err != nil {
				*s = ""
			}

			return true
		},
	},
	{
		idx: 13,
		check: func(s *string) bool {
			_, err := strconv.ParseFloat(*s, 32)
			if err != nil {
				*s = ""
			}

			return true
		},
	},
	{
		idx: 14,
		check: func(s *string) bool {
			_, err := strconv.ParseFloat(*s, 32)
			if err != nil {
				*s = ""
			}

			return true
		},
	},
	{
		idx: 15,
		check: func(s *string) bool {
			_, err := strconv.ParseFloat(*s, 32)
			if err != nil {
				*s = ""
			}

			return true
		},
	},
	{
		idx: 16,
		check: func(s *string) bool {
			_, err := strconv.ParseFloat(*s, 32)
			if err != nil {
				*s = ""
			}

			return true
		},
	},
	{
		idx: 17,
		check: func(s *string) bool {
			i, err := strconv.ParseFloat(*s, 32)
			if err != nil || i == 0 {
				return false
			}

			return true
		},
	},
	{
		idx: 18,
		check: func(s *string) bool {
			_, err := strconv.ParseFloat(*s, 32)
			if err != nil {
				*s = ""
			}

			return true
		},
	},
	{
		idx: 19,
		check: func(s *string) bool {
			_, err := strconv.ParseFloat(*s, 32)
			if err != nil {
				*s = ""
			}

			return true
		},
	},
}

type CrossColumnChecker struct {
	idxs  []int
	check func(*[]*string) bool
}

var crossColumnCheckers = []CrossColumnChecker{
	{
		idxs: []int{2, 3},
		check: func(cols *[]*string) bool {
			layoutDate := "2006-01-02 15:04:05"

			d1, d1err := time.Parse(layoutDate, *(*cols)[0])
			d2, d2err := time.Parse(layoutDate, *(*cols)[1])

			if d1err != nil || d2err != nil {
				return false
			}

			return d1.Compare(d2) == -1
		},
	},
	{
		idxs: []int{10, 11, 12, 13, 14, 15, 16, 17, 18, 19},
		check: func(cols *[]*string) bool {
			if *(*cols)[0] == "3" {
				for _, e := range (*cols)[1:] {
					if len(*e) == 0 {
						return true
					}
				}

				return false
			}

			return true
		},
	},
	{
		idxs: []int{11, 12, 13, 14, 15, 16, 17, 18, 19},
		check: func(cols *[]*string) bool {
			var sum float64 = 0
			for i := range *cols {
				if i != 6 {
					k, err := strconv.ParseFloat(*(*cols)[i], 32)
					if err != nil {
						return true
					}

					sum += k
				}
			}

			total, err := strconv.ParseFloat(*(*cols)[6], 32)
			if err != nil {
				return false
			}

			return total == sum
		},
	},
}
