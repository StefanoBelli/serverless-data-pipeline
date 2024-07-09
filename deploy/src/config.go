package main

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	apitypes "github.com/aws/aws-sdk-go-v2/service/apigatewayv2/types"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lmbdtypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/sfn"
)

/*
 * The following two constants may be changed without any repercussions
 */
const AWS_REGION = "us-east-1"
const IAM_ROLE = "LabRole"

/*
 * This is the definition of the critical pipeline state machine, format
 * specifiers are present to allow deployment code to replace them with
 * lambda function names at runtime.
 *
 * Any change to this JSON/AML specification should be done via Amazon's
 * visual editor
 */
const SFN_AML_DEFINITION_FMT = `
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
      "Next": "Was transformation possible?",
      "Catch": [
        {
          "ErrorEquals": [
            "States.TaskFailed"
          ],
          "Next": "Set validate tuple failed",
          "ResultPath": "$.error"
        }
      ]
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
      "Next": "Success",
      "Catch": [
        {
          "ErrorEquals": [
            "States.TaskFailed"
          ],
          "Next": "Parallel (1)",
          "ResultPath": "$.error"
        }
      ]
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

/*
 * DynamoDB tables
 * Changes in billing mode will not result in any kind of issues
 */
var tables = []dynamodb.CreateTableInput{
	{
		TableName: aws.String("validationStatus"),
		AttributeDefinitions: []ddbtypes.AttributeDefinition{
			{
				AttributeName: aws.String("StoreRequestId"),
				AttributeType: ddbtypes.ScalarAttributeTypeN,
			},
		},
		KeySchema: []ddbtypes.KeySchemaElement{
			{
				AttributeName: aws.String("StoreRequestId"),
				KeyType:       ddbtypes.KeyTypeHash,
			},
		},
		BillingMode: ddbtypes.BillingModePayPerRequest,
	},
	{
		TableName: aws.String("transformationStatus"),
		AttributeDefinitions: []ddbtypes.AttributeDefinition{
			{
				AttributeName: aws.String("StoreRequestId"),
				AttributeType: ddbtypes.ScalarAttributeTypeN,
			},
		},
		KeySchema: []ddbtypes.KeySchemaElement{
			{
				AttributeName: aws.String("StoreRequestId"),
				KeyType:       ddbtypes.KeyTypeHash,
			},
		},
		BillingMode: ddbtypes.BillingModePayPerRequest,
	},
	{
		TableName: aws.String("storeStatus"),
		AttributeDefinitions: []ddbtypes.AttributeDefinition{
			{
				AttributeName: aws.String("StoreRequestId"),
				AttributeType: ddbtypes.ScalarAttributeTypeN,
			},
		},
		KeySchema: []ddbtypes.KeySchemaElement{
			{
				AttributeName: aws.String("StoreRequestId"),
				KeyType:       ddbtypes.KeyTypeHash,
			},
		},
		BillingMode: ddbtypes.BillingModePayPerRequest,
	},
	{
		TableName: aws.String("nycYellowTaxis"),
		AttributeDefinitions: []ddbtypes.AttributeDefinition{
			{
				AttributeName: aws.String("StoreRequestId"),
				AttributeType: ddbtypes.ScalarAttributeTypeN,
			},
			{
				AttributeName: aws.String("EntryIdx"),
				AttributeType: ddbtypes.ScalarAttributeTypeN,
			},
		},
		KeySchema: []ddbtypes.KeySchemaElement{
			{
				AttributeName: aws.String("StoreRequestId"),
				KeyType:       ddbtypes.KeyTypeHash,
			},
			{
				AttributeName: aws.String("EntryIdx"),
				KeyType:       ddbtypes.KeyTypeRange, //sorting key
			},
		},
		BillingMode: ddbtypes.BillingModePayPerRequest,
	},
}

/*
 * Lambda functions to be coordinated via the state machine
 *
 * FunctionName and Timeout may be changed without any kind of issue.
 *
 * Changing the rest, might result in issues (golang compiler
 * target arch is x86_64, bootstrap is the name of the executable
 * which contains the main() and lambda handler functions)
 */
var lambdas = []lambda.CreateFunctionInput{
	{
		FunctionName:  aws.String("validate"),
		Role:          &iamRoleArn,
		PackageType:   lmbdtypes.PackageTypeZip,
		Architectures: []lmbdtypes.Architecture{lmbdtypes.ArchitectureX8664},
		Runtime:       lmbdtypes.RuntimeProvidedal2023,
		Handler:       aws.String("bootstrap"),
		Timeout:       aws.Int32(10),
	},
	{
		FunctionName:  aws.String("transform"),
		Role:          &iamRoleArn,
		PackageType:   lmbdtypes.PackageTypeZip,
		Architectures: []lmbdtypes.Architecture{lmbdtypes.ArchitectureX8664},
		Runtime:       lmbdtypes.RuntimeProvidedal2023,
		Handler:       aws.String("bootstrap"),
		Timeout:       aws.Int32(10),
	},
	{
		FunctionName:  aws.String("store"),
		Role:          &iamRoleArn,
		PackageType:   lmbdtypes.PackageTypeZip,
		Architectures: []lmbdtypes.Architecture{lmbdtypes.ArchitectureX8664},
		Runtime:       lmbdtypes.RuntimeProvidedal2023,
		Handler:       aws.String("bootstrap"),
		Timeout:       aws.Int32(10),
	},
	{
		FunctionName:  aws.String("flagValidateFailed"),
		Role:          &iamRoleArn,
		PackageType:   lmbdtypes.PackageTypeZip,
		Architectures: []lmbdtypes.Architecture{lmbdtypes.ArchitectureX8664},
		Runtime:       lmbdtypes.RuntimeProvidedal2023,
		Handler:       aws.String("bootstrap"),
		Timeout:       aws.Int32(10),
	},
	{
		FunctionName:  aws.String("flagTransformFailed"),
		Role:          &iamRoleArn,
		PackageType:   lmbdtypes.PackageTypeZip,
		Architectures: []lmbdtypes.Architecture{lmbdtypes.ArchitectureX8664},
		Runtime:       lmbdtypes.RuntimeProvidedal2023,
		Handler:       aws.String("bootstrap"),
		Timeout:       aws.Int32(10),
	},
	{
		FunctionName:  aws.String("flagStoreFailed"),
		Role:          &iamRoleArn,
		PackageType:   lmbdtypes.PackageTypeZip,
		Architectures: []lmbdtypes.Architecture{lmbdtypes.ArchitectureX8664},
		Runtime:       lmbdtypes.RuntimeProvidedal2023,
		Handler:       aws.String("bootstrap"),
		Timeout:       aws.Int32(10),
	},
}

/*
 * State machine name, may be changed without any kind of issue
 */
var stateMachine = sfn.CreateStateMachineInput{
	Name:    aws.String("CriticalDataPipeline"),
	RoleArn: &iamRoleArn,
}

/*
 * API gateway
 *
 * Name: may be changed without any kind of issue
 * ProtocolType: cannot be changed
 */
var api = apigatewayv2.CreateApiInput{
	Name:         aws.String("pipeline"),
	ProtocolType: apitypes.ProtocolTypeHttp,
}

/*
 * Do not touch any of the following lines
 */
var integration = apigatewayv2.CreateIntegrationInput{
	Description:          aws.String("CriticalDataPipeline integration"),
	IntegrationType:      apitypes.IntegrationTypeAwsProxy,
	IntegrationSubtype:   aws.String("StepFunctions-StartExecution"),
	PayloadFormatVersion: aws.String("1.0"),
	CredentialsArn:       &iamRoleArn,
	RequestParameters: map[string]string{
		"Input": "$request.body",
	},
}

/*
 * Do not touch any of the following lines
 */
var route = apigatewayv2.CreateRouteInput{
	RouteKey: aws.String("POST /store"),
}

/*
 * Name can be changed, do not change description
 */
var secret = secretsmanager.CreateSecretInput{
	Name:        aws.String("DataPipelineAuthKey"),
	Description: aws.String("secret for data pipeline HTTP methods"),
}

/*
 * Name may be changed, do not change other fields.
 */
var authorizer = apigatewayv2.CreateAuthorizerInput{
	Name:                           aws.String("DataPipelineAuthorizer"),
	AuthorizerType:                 apitypes.AuthorizerTypeRequest,
	IdentitySource:                 []string{"$request.header.Authorization"},
	AuthorizerPayloadFormatVersion: aws.String("2.0"),
	AuthorizerResultTtlInSeconds:   aws.Int32(0),
	EnableSimpleResponses:          aws.Bool(true),
	AuthorizerCredentialsArn:       &iamRoleArn,
}
