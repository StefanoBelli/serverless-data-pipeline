package flagPhaseFailed

/*
 * This package allows flagging your own support table entry
 * for failed transaction
 */

import (
	"context"
	"errors"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
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

// the first two fields are forwarded from the lambda
// JSON object answer (that lambda failed)
// The presence or not of the Error fields depends on
// which kind of error happened (see description in handler below)
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

// Recall that the state machine decides when it is needed to
// call flagValidateFailed, flagTransformFailed and flagStoreFailed.
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

	// attempting to update a non existant tuple results in error
	// which will be ignored (last step of the pipeline)
	_, err = dyndb.UpdateItem(dflCtx(), &dynamodb.UpdateItemInput{
		TableName:                 &tableName,
		Key:                       key,
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		UpdateExpression:          expr.Update(),
		ConditionExpression:       aws.String("attribute_exists(StoreRequestId)"),
	})

	return err
}

// Obtain a new DynamoDB client
func newDynamoDbService() (*dynamodb.Client, error) {
	// the "light" VM which runs this lambda has AWS_REGION env var set
	awsConfig, err := config.LoadDefaultConfig(
		dflCtx(),
		config.WithRegion(os.Getenv("AWS_REGION")))
	if err != nil {
		return nil, err
	}

	return dynamodb.NewFromConfig(awsConfig), nil
}

func getReasonCodeFromErrorType(errorType *string) int32 {
	// If the failure is caused by an error
	// Then it can only be treated if it was from transform or store
	// this is because, if validate errored then it would mean that
	// no entry was inserted in its DynamoDB support table,
	// the state machine stopped (no further phases were run, step function stopped)
	// and so no action can take place
	switch *errorType {
	case "TransformError":
		return 2
	case "StoreError":
		return 3
	}

	return 4 //Unknown
}

/* exported */

// Since flagStoreFailed, flagTransformFailed and flagValidateFailed are very similar
// it is worth to, instead of duplicating code, just make a "local module", generalize
// flagging of db by adding some little extra code and exporting two fundamental functions

// Handler function is called to handle a fail flagging request
func Handler(e failFlagRequest) (bool, error) {
	// This check is needed because the client needs to set the table name with
	// the exported function SetTableName, before calling Handler
	if tableNameIsSet {
		// Since this handler is being called when there is a fail in the pipeline
		// and the failure can be reported by either:
		//
		//   * the return value of a lambda (Reason is not 0 in JSON object response from the lambda),
		//     which naturally leads to different execution paths from the state machine
		//           e.Reason IS NOT ZERO
		//
		//   * the Golang "error" type which is treated by the state machine according
		//     to the runtime specification of the lambda (if it was Java/JVM, then it would be a
		//     "Throwable" object not being catched by the lambda code) and the error-handling policy
		//     of the state machine (what to do after an error of this kind, e.g. uncatched exception
		//     in the case of Java/JVM)
		//           e.Reason IS ZERO BUT the error type is attached
		//
		// If the "Reason" contained in the lambda response forwarded by the state machine to the
		// failure-handling code is 0 then it must be true that it is an error (meant as the second point
		// above) because the state machine invoked *THIS* lambda as a result of a failure
		// (repeating that a failure can be both determined via a decision block based on the result of
		// a lambda, or an error such as uncatched exception, which triggers other actions of the state machine)
		//
		// the store lambda failure will always result in the second point
		if e.Reason == 0 {
			// since the reason is 0, at this point, this means info about error is contained in
			// the "Error" field which must be present in the input JSON object passed to *THIS* lambda
			// from the state machine, following the errored condition from the previous lambda execution.
			// so we map it to an integer "Reason" which just denotes at which level of the
			// pipeline failure happened, this value is going to be updated in the transaction
			// entry in a DynamoDB table (specified by yourTableName set by the client code via exported SetTableName)
			e.Reason = getReasonCodeFromErrorType(&e.Error.Error)
		}

		// Get a new client for DynamoDB
		ddbSvc, err := newDynamoDbService()
		if err != nil {
			return false, err
		}

		// Update tuple (by its transaction id) to replace the reason from 0 to 1,2,3
		// 1 : validate failed
		// 2 : transform failed
		// 3 : store failed
		err = updateTuple(ddbSvc, e.TransactionId, e.Reason)
		return err == nil, err
	} else {
		// if table name is not set then ...
		return false, errors.New("need to set tableName (dev error)")
	}
}

// This function must be called before Handler()
// to set the DynamoDB table to be updated with
// reason code signaling fail conditions
func SetTableName(yourTableName string) {
	tableName = yourTableName
	tableNameIsSet = true
}
