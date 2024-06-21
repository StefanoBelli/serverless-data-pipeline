package main

import (
	"dyndbutils"
	"failsim"
	"fmt"
	"strconv"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
)

var FINAL_TABLE_NAME = "nycYellowTaxis"
var STATUS_TABLE_NAME = "storeStatus"

type TupleStoreRequest struct {
	Success       bool   `json:"success"`
	Reason        int    `json:"reason"`
	TransactionId uint64 `json:"transactionId"`
	Tuple         string `json:"tuple"`
}

type TupleStoreResponse struct {
	Success       bool   `json:"success"`
	Reason        int    `json:"reason"`
	TransactionId uint64 `json:"transactionId"`
}

type StoreError struct {
	cause   error
	userMsg string
}

func (se StoreError) Error() string {
	return fmt.Sprintf("%s: %v", se.userMsg, se.cause)
}

func erroredResponse(msg string, err error) (TupleStoreResponse, error) {
	return TupleStoreResponse{},
		StoreError{cause: err, userMsg: msg}
}

func validResponse() (TupleStoreResponse, error) {
	return TupleStoreResponse{Success: true}, nil
}

type NycYellowTaxiEntry struct {
	StoreRequestId       uint64  `dynamodbav:"StoreRequestId"`
	EntryIdx             int64   `dynamodbav:"EntryIdx"`
	VendorId             string  `dynamodbav:"VendorId"`
	PickupTime           string  `dynamodbav:"PickupTime"`
	DropoffTime          string  `dynamodbav:"DropoffTime"`
	PassengerCount       int64   `dynamodbav:"PassengerCount"`
	TripDistance         float64 `dynamodbav:"TripDistance"`
	RatecodeId           string  `dynamodbav:"RatecodeId"`
	StoreAndFwdFlag      string  `dynamodbav:"StoreAndFwdFlag"`
	PuLocationId         int64   `dynamodbav:"PuLocationId"`
	DoLocationId         int64   `dynamodbav:"DoLocationId"`
	PaymentType          string  `dynamodbav:"PaymentType"`
	FareAmount           float64 `dynamodbav:"FareAmount"`
	Extra                float64 `dynamodbav:"Extra"`
	MtaTax               float64 `dynamodbav:"MtaTax"`
	TipAmount            float64 `dynamodbav:"TipAmount"`
	TollsAmount          float64 `dynamodbav:"TollsAmount"`
	ImprovementSurcharge float64 `dynamodbav:"ImprovementSurcharge"`
	TotalAmount          float64 `dynamodbav:"TotalAmount"`
	CongestionSurcharge  float64 `dynamodbav:"CongestionSurcharge"`
	AirportFee           float64 `dynamodbav:"AirportFee"`
}

func parseDecInt64(from *string) (int64, error) {
	if len(*from) == 0 {
		return 0, nil
	}

	return strconv.ParseInt(*from, 10, 32)
}

func parseFloat64(from *string) (float64, error) {
	if len(*from) == 0 {
		return 0, nil
	}

	return strconv.ParseFloat(*from, 64)
}

func firstEncounteredError(errors *[]error) error {
	for _, err := range *errors {
		if err != nil {
			return err
		}
	}

	return nil
}

func populateEntryByRawTuple(entry *NycYellowTaxiEntry, id uint64, rawTuple *string) error {
	errs := make([]error, 14)

	fields := strings.Split(*rawTuple, "\t")

	entry.StoreRequestId = id
	entry.EntryIdx, errs[0] = parseDecInt64(&fields[0])
	entry.VendorId = fields[1]
	entry.PickupTime = fields[2]
	entry.DropoffTime = fields[3]
	entry.PassengerCount, errs[1] = parseDecInt64(&fields[4])
	entry.TripDistance, errs[2] = parseFloat64(&fields[5])
	entry.RatecodeId = fields[6]
	entry.StoreAndFwdFlag = fields[7]
	entry.PuLocationId, errs[3] = parseDecInt64(&fields[8])
	entry.DoLocationId, errs[4] = parseDecInt64(&fields[9])
	entry.PaymentType = fields[10]
	entry.FareAmount, errs[5] = parseFloat64(&fields[11])
	entry.Extra, errs[6] = parseFloat64(&fields[12])
	entry.MtaTax, errs[7] = parseFloat64(&fields[13])
	entry.TipAmount, errs[8] = parseFloat64(&fields[14])
	entry.TollsAmount, errs[9] = parseFloat64(&fields[15])
	entry.ImprovementSurcharge, errs[10] = parseFloat64(&fields[16])
	entry.TotalAmount, errs[11] = parseFloat64(&fields[17])
	entry.CongestionSurcharge, errs[12] = parseFloat64(&fields[18])
	entry.AirportFee, errs[13] = parseFloat64(&fields[19])

	return firstEncounteredError(&errs)
}

func handler(e TupleStoreRequest) (TupleStoreResponse, error) {
	ddbSvc, err := dyndbutils.NewDynamoDbService()
	if err != nil {
		return erroredResponse("unable to load dynamodb service", err)
	}

	// FAILSIM
	if err := failsim.OopsFailed(); err != nil {
		return erroredResponse("unable to put raw tuple", err)
	}
	// FAILSIM

	err = dyndbutils.PutInTable(
		ddbSvc,
		dyndbutils.BuildDefaultTupleStatus(e.TransactionId, &e.Tuple),
		&STATUS_TABLE_NAME)

	if err != nil {
		return erroredResponse("unable to put raw tuple", err)
	}

	nyte := NycYellowTaxiEntry{}

	err = populateEntryByRawTuple(&nyte, e.TransactionId, &e.Tuple)

	// FAILSIM
	if err == nil {
		err = failsim.OopsFailed()
	}
	// FAILSIM

	if err != nil {
		return erroredResponse("unable to populate entry from raw tuple", err)
	}

	//FAILSIM
	if err := failsim.OopsFailed(); err != nil {
		return erroredResponse("unable to put entry in final table", err)
	}
	//FAILSIM

	err = dyndbutils.PutInTable(
		ddbSvc,
		nyte,
		&FINAL_TABLE_NAME)

	if err != nil {
		return erroredResponse("unable to put entry in final table", err)
	}

	return validResponse()
}

func main() {
	lambda.Start(handler)
}
