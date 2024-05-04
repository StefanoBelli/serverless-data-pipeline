package main

import (
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	apitypes "github.com/aws/aws-sdk-go-v2/service/apigatewayv2/types"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lmbdtypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/aws/aws-sdk-go-v2/service/sfn"
)

const AWS_REGION = "us-east-1"

var IAM_LABROLE = "LabRole"

var SFN_AML_DEFINITION_FMT = `
{
  "Comment": "A description of my state machine",
  "StartAt": "Validate",
  "States": {
    "Validate": {
      "Type": "Task",
      "Resource": "arn:aws:states:::lambda:invoke",
      "OutputPath": "$.Payload",
      "Parameters": {
        "Payload.$": "$",
        "FunctionName": "%s"
      },
      "Retry": [
        {
          "ErrorEquals": [
            "Lambda.ServiceException",
            "Lambda.AWSLambdaException",
            "Lambda.SdkClientException",
            "Lambda.TooManyRequestsException"
          ],
          "IntervalSeconds": 1,
          "MaxAttempts": 3,
          "BackoffRate": 2
        }
      ],
      "Next": "Are validation checks passing?"
    },
    "Are validation checks passing?": {
      "Type": "Choice",
      "Choices": [
        {
          "Variable": "$.success",
          "BooleanEquals": true,
          "Next": "Transform"
        }
      ],
      "Default": "Set validate tuple failed"
    },
    "Transform": {
      "Type": "Task",
      "Resource": "arn:aws:states:::lambda:invoke",
      "OutputPath": "$.Payload",
      "Parameters": {
        "Payload.$": "$",
        "FunctionName": "%s"
      },
      "Retry": [
        {
          "ErrorEquals": [
            "Lambda.ServiceException",
            "Lambda.AWSLambdaException",
            "Lambda.SdkClientException",
            "Lambda.TooManyRequestsException"
          ],
          "IntervalSeconds": 1,
          "MaxAttempts": 3,
          "BackoffRate": 2
        }
      ],
      "Next": "Was transformation possible?"
    },
    "Was transformation possible?": {
      "Type": "Choice",
      "Choices": [
        {
          "Variable": "$.success",
          "BooleanEquals": true,
          "Next": "Store"
        }
      ],
      "Default": "Parallel"
    },
    "Parallel": {
      "Type": "Parallel",
      "Next": "Fail - TransformFailure",
      "Branches": [
        {
          "StartAt": "Set transform tuple failed",
          "States": {
            "Set transform tuple failed": {
              "Type": "Task",
              "Resource": "arn:aws:states:::lambda:invoke",
              "OutputPath": "$.Payload",
              "Parameters": {
                "Payload.$": "$",
                "FunctionName": "%s"
              },
              "Retry": [
                {
                  "ErrorEquals": [
                    "Lambda.ServiceException",
                    "Lambda.AWSLambdaException",
                    "Lambda.SdkClientException",
                    "Lambda.TooManyRequestsException"
                  ],
                  "IntervalSeconds": 1,
                  "MaxAttempts": 3,
                  "BackoffRate": 2
                }
              ],
              "End": true
            }
          }
        },
        {
          "StartAt": "Set validate tuple failed (1)",
          "States": {
            "Set validate tuple failed (1)": {
              "Type": "Task",
              "Resource": "arn:aws:states:::lambda:invoke",
              "OutputPath": "$.Payload",
              "Parameters": {
                "Payload.$": "$",
                "FunctionName": "%s"
              },
              "Retry": [
                {
                  "ErrorEquals": [
                    "Lambda.ServiceException",
                    "Lambda.AWSLambdaException",
                    "Lambda.SdkClientException",
                    "Lambda.TooManyRequestsException"
                  ],
                  "IntervalSeconds": 1,
                  "MaxAttempts": 3,
                  "BackoffRate": 2
                }
              ],
              "End": true
            }
          }
        }
      ]
    },
    "Store": {
      "Type": "Task",
      "Resource": "arn:aws:states:::lambda:invoke",
      "OutputPath": "$.Payload",
      "Parameters": {
        "Payload.$": "$",
        "FunctionName": "%s"
      },
      "Retry": [
        {
          "ErrorEquals": [
            "Lambda.ServiceException",
            "Lambda.AWSLambdaException",
            "Lambda.SdkClientException",
            "Lambda.TooManyRequestsException"
          ],
          "IntervalSeconds": 1,
          "MaxAttempts": 3,
          "BackoffRate": 2
        }
      ],
      "Next": "Did the storage process encounter any error?"
    },
    "Did the storage process encounter any error?": {
      "Type": "Choice",
      "Choices": [
        {
          "Variable": "$.success",
          "BooleanEquals": true,
          "Next": "Success"
        }
      ],
      "Default": "Parallel (1)"
    },
    "Parallel (1)": {
      "Type": "Parallel",
      "Next": "Fail - StoreFailure",
      "Branches": [
        {
          "StartAt": "Set store tuple failed",
          "States": {
            "Set store tuple failed": {
              "Type": "Task",
              "Resource": "arn:aws:states:::lambda:invoke",
              "OutputPath": "$.Payload",
              "Parameters": {
                "Payload.$": "$",
                "FunctionName": "%s"
              },
              "Retry": [
                {
                  "ErrorEquals": [
                    "Lambda.ServiceException",
                    "Lambda.AWSLambdaException",
                    "Lambda.SdkClientException",
                    "Lambda.TooManyRequestsException"
                  ],
                  "IntervalSeconds": 1,
                  "MaxAttempts": 3,
                  "BackoffRate": 2
                }
              ],
              "End": true
            }
          }
        },
        {
          "StartAt": "Set transform tuple failed (1)",
          "States": {
            "Set transform tuple failed (1)": {
              "Type": "Task",
              "Resource": "arn:aws:states:::lambda:invoke",
              "OutputPath": "$.Payload",
              "Parameters": {
                "Payload.$": "$",
                "FunctionName": "%s"
              },
              "Retry": [
                {
                  "ErrorEquals": [
                    "Lambda.ServiceException",
                    "Lambda.AWSLambdaException",
                    "Lambda.SdkClientException",
                    "Lambda.TooManyRequestsException"
                  ],
                  "IntervalSeconds": 1,
                  "MaxAttempts": 3,
                  "BackoffRate": 2
                }
              ],
              "End": true
            }
          }
        },
        {
          "StartAt": "Set validate tuple failed (2)",
          "States": {
            "Set validate tuple failed (2)": {
              "Type": "Task",
              "Resource": "arn:aws:states:::lambda:invoke",
              "OutputPath": "$.Payload",
              "Parameters": {
                "Payload.$": "$",
                "FunctionName": "%s"
              },
              "Retry": [
                {
                  "ErrorEquals": [
                    "Lambda.ServiceException",
                    "Lambda.AWSLambdaException",
                    "Lambda.SdkClientException",
                    "Lambda.TooManyRequestsException"
                  ],
                  "IntervalSeconds": 1,
                  "MaxAttempts": 3,
                  "BackoffRate": 2
                }
              ],
              "End": true
            }
          }
        }
      ]
    },
    "Success": {
      "Type": "Succeed"
    },
    "Fail - StoreFailure": {
      "Type": "Fail"
    },
    "Fail - TransformFailure": {
      "Type": "Fail"
    },
    "Set validate tuple failed": {
      "Type": "Task",
      "Resource": "arn:aws:states:::lambda:invoke",
      "OutputPath": "$.Payload",
      "Parameters": {
        "Payload.$": "$",
        "FunctionName": "%s"
      },
      "Retry": [
        {
          "ErrorEquals": [
            "Lambda.ServiceException",
            "Lambda.AWSLambdaException",
            "Lambda.SdkClientException",
            "Lambda.TooManyRequestsException"
          ],
          "IntervalSeconds": 1,
          "MaxAttempts": 3,
          "BackoffRate": 2
        }
      ],
      "Next": "Fail - ValidateFailure"
    },
    "Fail - ValidateFailure": {
      "Type": "Fail"
    }
  }
}
`

var dynamoDbTables = []dynamodb.CreateTableInput{
	{
		TableName: &validationStatus,
		AttributeDefinitions: []ddbtypes.AttributeDefinition{
			{
				AttributeName: &idx,
				AttributeType: ddbtypes.ScalarAttributeTypeN,
			},
			{
				AttributeName: &nextIdx,
				AttributeType: ddbtypes.ScalarAttributeTypeN,
			},
			{
				AttributeName: &origData,
				AttributeType: ddbtypes.ScalarAttributeTypeS,
			},
			{
				AttributeName: &status,
				AttributeType: ddbtypes.ScalarAttributeTypeN,
			},
			{
				AttributeName: &reason,
				AttributeType: ddbtypes.ScalarAttributeTypeS,
			},
		},
		KeySchema: []ddbtypes.KeySchemaElement{
			{
				AttributeName: &idx,
				KeyType:       ddbtypes.KeyTypeHash,
			},
		},
		BillingMode: ddbtypes.BillingModePayPerRequest,
	},
	{
		TableName: &transformationStatus,
		AttributeDefinitions: []ddbtypes.AttributeDefinition{
			{
				AttributeName: &idx,
				AttributeType: ddbtypes.ScalarAttributeTypeN,
			},
			{
				AttributeName: &nextIdx,
				AttributeType: ddbtypes.ScalarAttributeTypeN,
			},
			{
				AttributeName: &origData,
				AttributeType: ddbtypes.ScalarAttributeTypeS,
			},
			{
				AttributeName: &status,
				AttributeType: ddbtypes.ScalarAttributeTypeN,
			},
			{
				AttributeName: &reason,
				AttributeType: ddbtypes.ScalarAttributeTypeS,
			},
		},
		KeySchema: []ddbtypes.KeySchemaElement{
			{
				AttributeName: &idx,
				KeyType:       ddbtypes.KeyTypeHash,
			},
		},
		BillingMode: ddbtypes.BillingModePayPerRequest,
	},
	{
		TableName: &storeStatus,
		AttributeDefinitions: []ddbtypes.AttributeDefinition{
			{
				AttributeName: &idx,
				AttributeType: ddbtypes.ScalarAttributeTypeN,
			},
			{
				AttributeName: &origData,
				AttributeType: ddbtypes.ScalarAttributeTypeS,
			},
			{
				AttributeName: &status,
				AttributeType: ddbtypes.ScalarAttributeTypeN,
			},
			{
				AttributeName: &reason,
				AttributeType: ddbtypes.ScalarAttributeTypeS,
			},
		},
		KeySchema: []ddbtypes.KeySchemaElement{
			{
				AttributeName: &idx,
				KeyType:       ddbtypes.KeyTypeHash,
			},
		},
		BillingMode: ddbtypes.BillingModePayPerRequest,
	},
	{
		TableName: &nycYellowTaxis,
		AttributeDefinitions: []ddbtypes.AttributeDefinition{
			{
				AttributeName: &idx,
				AttributeType: ddbtypes.ScalarAttributeTypeN,
			},
			{
				AttributeName: &vendor,
				AttributeType: ddbtypes.ScalarAttributeTypeS,
			},
			{
				AttributeName: &pickupTime,
				AttributeType: ddbtypes.ScalarAttributeTypeS,
			},
			{
				AttributeName: &dropoffTime,
				AttributeType: ddbtypes.ScalarAttributeTypeS,
			},
			{
				AttributeName: &passengerCount,
				AttributeType: ddbtypes.ScalarAttributeTypeN,
			},
			{
				AttributeName: &tripDistance,
				AttributeType: ddbtypes.ScalarAttributeTypeN,
			},
			{
				AttributeName: &ratecode,
				AttributeType: ddbtypes.ScalarAttributeTypeS,
			},
			{
				AttributeName: &storeAndFwdFlag,
				AttributeType: ddbtypes.ScalarAttributeTypeB,
			},
			{
				AttributeName: &puLocationId,
				AttributeType: ddbtypes.ScalarAttributeTypeN,
			},
			{
				AttributeName: &doLocationId,
				AttributeType: ddbtypes.ScalarAttributeTypeN,
			},
			{
				AttributeName: &paymentType,
				AttributeType: ddbtypes.ScalarAttributeTypeS,
			},
			{
				AttributeName: &fareAmount,
				AttributeType: ddbtypes.ScalarAttributeTypeN,
			},
			{
				AttributeName: &extra,
				AttributeType: ddbtypes.ScalarAttributeTypeN,
			},
			{
				AttributeName: &mtaTax,
				AttributeType: ddbtypes.ScalarAttributeTypeN,
			},
			{
				AttributeName: &tipAmount,
				AttributeType: ddbtypes.ScalarAttributeTypeN,
			},
			{
				AttributeName: &tollsAmount,
				AttributeType: ddbtypes.ScalarAttributeTypeN,
			},
			{
				AttributeName: &improvementSurcharge,
				AttributeType: ddbtypes.ScalarAttributeTypeN,
			},
			{
				AttributeName: &totalAmount,
				AttributeType: ddbtypes.ScalarAttributeTypeN,
			},
			{
				AttributeName: &congestionSurcharge,
				AttributeType: ddbtypes.ScalarAttributeTypeN,
			},
			{
				AttributeName: &airportFee,
				AttributeType: ddbtypes.ScalarAttributeTypeN,
			},
		},
		KeySchema: []ddbtypes.KeySchemaElement{
			{
				AttributeName: &idx,
				KeyType:       ddbtypes.KeyTypeHash,
			},
		},
		BillingMode: ddbtypes.BillingModePayPerRequest,
	},
}

var lambdaFunctions = []lambda.CreateFunctionInput{
	{
		FunctionName:  &validate,
		Code:          &lmbdtypes.FunctionCode{ZipFile: nil},
		PackageType:   lmbdtypes.PackageTypeZip,
		Role:          iamLabRole.Arn,
		Architectures: []lmbdtypes.Architecture{lmbdtypes.ArchitectureX8664},
		Runtime:       lmbdtypes.RuntimeGo1x,
		Timeout:       &lambdaTimeout,
	},
	{
		FunctionName:  &transform,
		Code:          &lmbdtypes.FunctionCode{ZipFile: nil},
		PackageType:   lmbdtypes.PackageTypeZip,
		Role:          iamLabRole.Arn,
		Architectures: []lmbdtypes.Architecture{lmbdtypes.ArchitectureX8664},
		Runtime:       lmbdtypes.RuntimeGo1x,
		Timeout:       &lambdaTimeout,
	},
	{
		FunctionName:  &store,
		Code:          &lmbdtypes.FunctionCode{ZipFile: nil},
		PackageType:   lmbdtypes.PackageTypeZip,
		Role:          iamLabRole.Arn,
		Architectures: []lmbdtypes.Architecture{lmbdtypes.ArchitectureX8664},
		Runtime:       lmbdtypes.RuntimeGo1x,
		Timeout:       &lambdaTimeout,
	},
	{
		FunctionName:  &flagValidateFailed,
		Code:          &lmbdtypes.FunctionCode{ZipFile: nil},
		PackageType:   lmbdtypes.PackageTypeZip,
		Role:          iamLabRole.Arn,
		Architectures: []lmbdtypes.Architecture{lmbdtypes.ArchitectureX8664},
		Runtime:       lmbdtypes.RuntimeGo1x,
		Timeout:       &lambdaTimeout,
	},
	{
		FunctionName:  &flagTransformFailed,
		Code:          &lmbdtypes.FunctionCode{ZipFile: nil},
		PackageType:   lmbdtypes.PackageTypeZip,
		Role:          iamLabRole.Arn,
		Architectures: []lmbdtypes.Architecture{lmbdtypes.ArchitectureX8664},
		Runtime:       lmbdtypes.RuntimeGo1x,
		Timeout:       &lambdaTimeout,
	},
	{
		FunctionName:  &flagStoreFailed,
		Code:          &lmbdtypes.FunctionCode{ZipFile: nil},
		PackageType:   lmbdtypes.PackageTypeZip,
		Role:          iamLabRole.Arn,
		Architectures: []lmbdtypes.Architecture{lmbdtypes.ArchitectureX8664},
		Runtime:       lmbdtypes.RuntimeGo1x,
		Timeout:       &lambdaTimeout,
	},
}

var stateMachine = sfn.CreateStateMachineInput{
	Name:    &criticalDataPipeline,
	RoleArn: iamLabRole.Arn,
}

var api = apigatewayv2.CreateApiInput{
	Name:         &pipeline,
	ProtocolType: apitypes.ProtocolTypeHttp,
}

var integrations = []apigatewayv2.CreateIntegrationInput{
	{
		Description:       &criticalDataPipelineIntegration,
		IntegrationMethod: &post,
	},
}

var routes = []apigatewayv2.CreateRouteInput{
	{
		RouteKey: &slashStore,
		//AuthorizationScopes:
		//AuthorizationType:
		//AuthorizerId:
	},
}
