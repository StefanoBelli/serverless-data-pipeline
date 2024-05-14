package main

import (
	"dyndbutils"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
)

var TABLE_NAME = "transformationStatus"

type TupleTransformationResponse struct {
	Success         bool   `json:"success"`
	Reason          int    `json:"reason"`
	TransactionUuid int64  `json:"transactionUuid"`
	Tuple           string `json:"tuple"`
}

type TupleTransformationRequest struct {
	TransactionUuid int64  `json:"transactionUuid"`
	Tuple           string `json:"tuple"`
}

type TransformError struct {
	cause   error
	userMsg string
}

func (te TransformError) Error() string {
	return fmt.Sprintf("%s: %v", te.userMsg, te.cause)
}

func erroredResponse(msg string, err error) (TupleTransformationResponse, error) {
	return TupleTransformationResponse{},
		TransformError{cause: err, userMsg: msg}
}

func validResponse(e *TupleTransformationRequest) (TupleTransformationResponse, error) {
	return TupleTransformationResponse{
		Success:         true,
		Reason:          0,
		TransactionUuid: e.TransactionUuid,
		Tuple:           e.Tuple,
	}, nil
}

func invalidResponse(e *TupleTransformationRequest) (TupleTransformationResponse, error) {
	return TupleTransformationResponse{
		Success:         false,
		Reason:          2,
		TransactionUuid: e.TransactionUuid,
		Tuple:           e.Tuple,
	}, nil
}

/*
 * According to: https://www.nyc.gov/assets/tlc/downloads/pdf/data_dictionary_trip_records_yellow.pdf
 * Transform:
 *  - VendorID
 *  - RateCodeID
 *  - Store_and_fwd_flag
 *  - Payment_type
 *
 * Also transform:
 *  - CSV separator character
 *  - Date format
 *  - Passenger_count from float to int
 *
 * Transform fails for:
 *  - negative USD
 *  - unknown mapping from code to mnemonic value (striclty follows PDF above)
 */
func performDataTransformation(rawTuple *string) bool {
	const oldCsvCommaSep = ","
	const newCsvCommaSep = "\t"

	csvCols := strings.Split(*rawTuple, oldCsvCommaSep)

	for _, columns := range multiColumnTransformers {
		var cols []*string
		for _, ei := range columns.idxs {
			cols = append(cols, &csvCols[ei])
		}

		if !columns.transform(&cols) {
			return false
		}
	}

	//also rejoin by transforming sep. csv char
	*rawTuple = strings.Join(csvCols, newCsvCommaSep)
	return true
}

func handler(e TupleTransformationRequest) (TupleTransformationResponse, error) {
	ddbSvc, err := dyndbutils.NewDynamoDbService()
	if err != nil {
		return erroredResponse("unable to load dynamodb service", err)
	}

	err = dyndbutils.PutInTable(
		ddbSvc,
		dyndbutils.BuildDefaultTupleStatus(e.TransactionUuid, &e.Tuple),
		&TABLE_NAME)
	if err != nil {
		return erroredResponse("unable to put raw tuple", err)
	}

	if performDataTransformation(&e.Tuple) {
		return validResponse(&e)
	} else {
		return invalidResponse(&e)
	}
}

func main() {
	lambda.Start(handler)
}

type MultiColumnTransformer struct {
	idxs      []int
	transform func(*[]*string) bool
}

var multiColumnTransformers = []MultiColumnTransformer{
	{
		idxs: []int{1},
		transform: func(col *[]*string) bool {
			assoc := map[string]string{
				"1": "Creative Mobile Technologies, LLC",
				"2": "VeriFone Inc.",
			}

			return applySubst((*col)[0], &assoc)
		},
	},
	{
		idxs: []int{4},
		transform: func(col *[]*string) bool {
			fl, err := strconv.ParseFloat(*(*col)[0], 64)
			if err != nil {
				return false
			}

			*(*col)[0] = fmt.Sprintf("%d", int(fl))
			return true
		},
	},
	{
		idxs: []int{6},
		transform: func(col *[]*string) bool {
			assoc := map[string]string{
				"1.0":  "Standard rate",
				"2.0":  "JFK",
				"3.0":  "Newark",
				"4.0":  "Nassau or Westchester",
				"5.0":  "Negotiated fare",
				"6.0":  "Group ride",
				"99.0": "Unknown (type=99)",
			}

			return applySubst((*col)[0], &assoc)
		},
	},
	{
		idxs: []int{7},
		transform: func(col *[]*string) bool {
			assoc := map[string]string{
				"Y": "store and forward trip",
				"N": "not a store and forward trip",
			}

			return applySubst((*col)[0], &assoc)
		},
	},
	{
		idxs: []int{10},
		transform: func(col *[]*string) bool {
			assoc := map[string]string{
				"1": "Credit card",
				"2": "Cash",
				"3": "No charge",
				"4": "Dispute",
				"5": "Unknown",
				"6": "Voided trip",
			}

			return applySubst((*col)[0], &assoc)
		},
	},
	{
		idxs: []int{2, 3},
		transform: func(cols *[]*string) bool {
			for _, col := range *cols {
				dt, err := time.Parse("2006-01-02 15:04:05", *col)
				if err != nil {
					return false
				}

				*col = dt.Format("02/01/2006 15:04:05")
			}

			return true
		},
	},
}

func applySubst(target *string, m *map[string]string) bool {
	found := false

	for k := range *m {
		if *target == k {
			found = true
			break
		}
	}

	if found {
		*target = (*m)[*target]
	}

	return found
}
