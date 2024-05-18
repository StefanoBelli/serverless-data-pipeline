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
	return strconv.ParseInt(*from, 10, 32)
}

func parseFloat64(from *string) (float64, error) {
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
	errs := make([]error, 13)

	fields := strings.Split(*rawTuple, "\t")

	entry.StoreRequestId = id
	entry.VendorId = fields[1]
	entry.PickupTime = fields[2]
	entry.DropoffTime = fields[3]
	entry.PassengerCount, errs[0] = parseDecInt64(&fields[4])
	entry.TripDistance, errs[1] = parseFloat64(&fields[5])
	entry.RatecodeId = fields[6]
	entry.StoreAndFwdFlag = fields[7]
	entry.PuLocationId, errs[2] = parseDecInt64(&fields[8])
	entry.DoLocationId, errs[3] = parseDecInt64(&fields[9])
	entry.PaymentType = fields[10]
	entry.FareAmount, errs[4] = parseFloat64(&fields[11])
	entry.Extra, errs[5] = parseFloat64(&fields[12])
	entry.MtaTax, errs[6] = parseFloat64(&fields[13])
	entry.TipAmount, errs[7] = parseFloat64(&fields[14])
	entry.TollsAmount, errs[8] = parseFloat64(&fields[15])
	entry.ImprovementSurcharge, errs[9] = parseFloat64(&fields[16])
	entry.TotalAmount, errs[10] = parseFloat64(&fields[17])
	entry.CongestionSurcharge, errs[11] = parseFloat64(&fields[18])
	entry.AirportFee, errs[12] = parseFloat64(&fields[19])

	return firstEncounteredError(&errs)
}

func handler(e TupleStoreRequest) (TupleStoreResponse, error) {
	ddbSvc, err := dyndbutils.NewDynamoDbService()

	if err == nil {
		err = failsim.OopsFailed()
	}

	if err != nil {
		return erroredResponse("unable to load dynamodb service", err)
	}

	err = dyndbutils.PutInTable(
		ddbSvc,
		dyndbutils.BuildDefaultTupleStatus(e.TransactionId, &e.Tuple),
		&STATUS_TABLE_NAME)

	if err == nil {
		err = failsim.OopsFailed()
	}

	if err != nil {
		return erroredResponse("unable to put raw tuple", err)
	}

	nyte := NycYellowTaxiEntry{}

	err = populateEntryByRawTuple(&nyte, e.TransactionId, &e.Tuple)

	if err == nil {
		err = failsim.OopsFailed()
	}

	if err != nil {
		return erroredResponse("unable to populate entry from raw tuple", err)
	}

	err = dyndbutils.PutInTable(
		ddbSvc,
		nyte,
		&FINAL_TABLE_NAME)

	if err == nil {
		err = failsim.OopsFailed()
	}

	if err != nil {
		return erroredResponse("unable to put entry in final table", err)
	}

	return validResponse()
}

func main() {
	lambda.Start(handler)
}
