package main

import (
	"context"
	"errors"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
)

func dflCtx() context.Context {
	return context.TODO()
}

// get a new secrets manager client
func newSecretsManagerService() (*secretsmanager.Client, error) {
	awsConfig, err := config.LoadDefaultConfig(
		dflCtx(),
		config.WithRegion(os.Getenv("AWS_REGION")))
	if err != nil {
		return nil, err
	}

	return secretsmanager.NewFromConfig(awsConfig), nil
}

// tell secretsmanager to decrypt secret
func getSecretValue() (string, error) {
	smSvc, err := newSecretsManagerService()
	if err != nil {
		return "", err
	}

	// We will need the ARN of the cryptographic storage to
	// be able to use it
	// IMPORTANT: name in Values array needs to be changed if
	//            secret storage name changes (see ../../deploy/src/config.go)
	lsi := secretsmanager.ListSecretsInput{
		SortOrder:  types.SortOrderTypeDesc,
		MaxResults: aws.Int32(1),
		Filters: []types.Filter{
			{
				Key:    types.FilterNameStringTypeName,
				Values: []string{"DataPipelineAuthKey"},
			},
		},
	}
	lso, err := smSvc.ListSecrets(dflCtx(), &lsi)
	if err != nil {
		return "", err
	}

	if len(lso.SecretList) == 0 {
		return "", errors.New("SecretList is empty")
	}

	// Got the crypto-storage ARN, so we obtain the plaintext secret
	gsvi := secretsmanager.GetSecretValueInput{
		SecretId: lso.SecretList[0].ARN,
	}
	gsvo, err := smSvc.GetSecretValue(dflCtx(), &gsvi)
	if err != nil {
		return "", err
	}

	return string(gsvo.SecretBinary), nil
}

type AuthorizationResponse struct {
	IsAuthorized bool `json:"isAuthorized"`
}

// Main lambda handler
func handler(e events.APIGatewayV2CustomAuthorizerV2Request) (AuthorizationResponse, error) {
	realKey, err := getSecretValue()
	if err != nil {
		return AuthorizationResponse{IsAuthorized: false}, err
	}

	// Check the plaintext key obtained via the secretsmanager against
	// the one sent via HTTP (encrypted) header field "Authorization"
	// by the client, JSON object answer will allow Authorizer to determine
	// if allow or deny resource access
	return AuthorizationResponse{
		IsAuthorized: realKey == e.IdentitySource[0],
	}, nil
}

func main() {
	lambda.Start(handler)
}
