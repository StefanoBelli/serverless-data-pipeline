package main

import (
	"dyndbutils"
	"errors"
	"failsim"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
)

const CSV_COMMA_SEP = ","
const CSV_NUMCOLS = 20

var TABLE_NAME = "validationStatus"

type TupleValidationRequest struct {
	Tuple string `json:"tuple"`
}

type TupleValidationResponse struct {
	Success       bool   `json:"success"`
	Reason        int    `json:"reason"`
	TransactionId uint64 `json:"transactionId"`
	Tuple         string `json:"tuple"`
}

func validResponse(id uint64, rawTuple *string) (TupleValidationResponse, error) {
	return TupleValidationResponse{
		Success:       true,
		Reason:        0,
		TransactionId: id,
		Tuple:         *rawTuple,
	}, nil
}

func invalidResponse(id uint64) (TupleValidationResponse, error) {
	return TupleValidationResponse{
		Success:       false,
		Reason:        1,
		TransactionId: id,
	}, nil
}

func erroredResponse(msg string, err error) (TupleValidationResponse, error) {
	return TupleValidationResponse{}, fmt.Errorf("%s: %v", msg, err)
}

func calculateTransactionId(rawTuple *string, unix32BeginTime int64) uint64 {
	var id uint64 = uint64(unix32BeginTime) + uint64(rand.Intn(10000))
	for i, c := range *rawTuple {
		id += uint64(c) * uint64(i)
	}

	return id
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

	*tuple = strings.Join(csvCols, ",")

	return true
}

func handler(e TupleValidationRequest) (TupleValidationResponse, error) {
	transactionBeginTime := time.Now().Unix()

	fixedTuple := strings.TrimSpace(e.Tuple)
	if len(fixedTuple) == 0 {
		return erroredResponse("receiving input",
			errors.New("empty tuple"))
	}

	ddbSvc, err := dyndbutils.NewDynamoDbService()
	if err != nil {
		return erroredResponse("unable to load dynamodb service", err)
	}

	transactionId := calculateTransactionId(&e.Tuple, transactionBeginTime)

	// FAILSIM
	if err := failsim.OopsFailed(); err != nil {
		return erroredResponse("unable to put raw table", err)
	}
	// FAILSIM

	err = dyndbutils.PutInTable(
		ddbSvc,
		dyndbutils.BuildDefaultTupleStatus(transactionId, &e.Tuple),
		&TABLE_NAME)

	if err != nil {
		return erroredResponse("unable to put raw tuple", err)
	}

	if fieldChecksAreOk(&fixedTuple) {
		return validResponse(transactionId, &fixedTuple)
	} else {
		return invalidResponse(transactionId)
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
			if err != nil || i < 1 || i > 5 {
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
			if err != nil || i < 1 || (i > 6 && i != 99) {
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
			if err != nil || i < 1 || i > 6 {
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

			return d1.Before(d2)
		},
	},
	{
		idxs: []int{11, 12, 13, 14, 15, 16, 17, 18, 19},
		check: func(cols *[]*string) bool {
			var sum float64 = 0
			for i := range *cols {
				if i != 6 {
					k, err := strconv.ParseFloat(*(*cols)[i], 64)
					if err != nil {
						return true
					}

					sum += k
				}
			}

			total, err := strconv.ParseFloat(*(*cols)[6], 64)
			if err != nil {
				return false
			}

			return total == sum
		},
	},
}
