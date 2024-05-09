package main

import (
	"context"
	"errors"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
)

func dflCtx() context.Context {
	return context.TODO()
}

func newSecretsManagerService() (*secretsmanager.Client, error) {
	awsConfig, err := config.LoadDefaultConfig(
		dflCtx(),
		config.WithRegion(os.Getenv("AWS_REGION")))
	if err != nil {
		return nil, err
	}

	return secretsmanager.NewFromConfig(awsConfig), nil
}

func getSecretValue() (string, error) {
	var i1 int32 = 1

	smSvc, err := newSecretsManagerService()
	if err != nil {
		return "", err
	}

	lsi := secretsmanager.ListSecretsInput{
		SortOrder:  types.SortOrderTypeDesc,
		MaxResults: &i1,
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

func handler(e events.APIGatewayV2CustomAuthorizerV2Request) (AuthorizationResponse, error) {
	realKey, err := getSecretValue()
	if err != nil {
		return AuthorizationResponse{IsAuthorized: false}, err
	}

	return AuthorizationResponse{
		IsAuthorized: realKey == e.IdentitySource[0],
	}, nil
}

func main() {
	lambda.Start(handler)
}
