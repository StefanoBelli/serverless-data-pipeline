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

var lambdas = []lambda.CreateFunctionInput{
	{
		FunctionName:  &validate,
		PackageType:   lmbdtypes.PackageTypeZip,
		Architectures: []lmbdtypes.Architecture{lmbdtypes.ArchitectureX8664},
		Runtime:       lmbdtypes.RuntimeProvidedal2023,
		Handler:       &bootstrap,
		Timeout:       &lambdaTimeout,
	},
	{
		FunctionName:  &transform,
		PackageType:   lmbdtypes.PackageTypeZip,
		Architectures: []lmbdtypes.Architecture{lmbdtypes.ArchitectureX8664},
		Runtime:       lmbdtypes.RuntimeProvidedal2023,
		Handler:       &bootstrap,
		Timeout:       &lambdaTimeout,
	},
	{
		FunctionName:  &store,
		PackageType:   lmbdtypes.PackageTypeZip,
		Architectures: []lmbdtypes.Architecture{lmbdtypes.ArchitectureX8664},
		Runtime:       lmbdtypes.RuntimeProvidedal2023,
		Handler:       &bootstrap,
		Timeout:       &lambdaTimeout,
	},
	{
		FunctionName:  &flagValidateFailed,
		PackageType:   lmbdtypes.PackageTypeZip,
		Architectures: []lmbdtypes.Architecture{lmbdtypes.ArchitectureX8664},
		Runtime:       lmbdtypes.RuntimeProvidedal2023,
		Handler:       &bootstrap,
		Timeout:       &lambdaTimeout,
	},
	{
		FunctionName:  &flagTransformFailed,
		PackageType:   lmbdtypes.PackageTypeZip,
		Architectures: []lmbdtypes.Architecture{lmbdtypes.ArchitectureX8664},
		Runtime:       lmbdtypes.RuntimeProvidedal2023,
		Handler:       &bootstrap,
		Timeout:       &lambdaTimeout,
	},
	{
		FunctionName:  &flagStoreFailed,
		PackageType:   lmbdtypes.PackageTypeZip,
		Architectures: []lmbdtypes.Architecture{lmbdtypes.ArchitectureX8664},
		Runtime:       lmbdtypes.RuntimeProvidedal2023,
		Handler:       &bootstrap,
		Timeout:       &lambdaTimeout,
	},
}

var stateMachine = sfn.CreateStateMachineInput{
	Name: &criticalDataPipeline,
}

var api = apigatewayv2.CreateApiInput{
	Name:         &pipeline,
	ProtocolType: apitypes.ProtocolTypeHttp,
}

var integrations = []apigatewayv2.CreateIntegrationInput{
	{
		Description:          &criticalDataPipelineIntegration,
		IntegrationType:      apitypes.IntegrationTypeAwsProxy,
		IntegrationSubtype:   &startExecution,
		PayloadFormatVersion: &twoDotZero,
	},
}

var routes = []apigatewayv2.CreateRouteInput{
	{
		RouteKey: &postSlashStore,
		//AuthorizationScopes:
		//AuthorizationType:
		//AuthorizerId:
	},
}
