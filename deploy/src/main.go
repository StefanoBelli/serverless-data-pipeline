package main

import (
	"context"
	"flag"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/sfn"
)

var awsConfig aws.Config
var iamLabRole iamtypes.Role

func createDynamoDbs() {
	dynamoDbSvc := dynamodb.NewFromConfig(awsConfig)

	for _, dbs := range dynamoDbTables {
		opOut, err := dynamoDbSvc.CreateTable(
			context.TODO(),
			&dbs)
		if err != nil {
			log.Fatalf("unable to create table: %v", err)
		}

		log.Printf(
			"table %s, arn %s, status %s\n",
			*opOut.TableDescription.TableName,
			*opOut.TableDescription.TableArn,
			opOut.TableDescription.TableStatus)
	}
}

func createLambdas() {
	lambdaSvc := lambda.NewFromConfig(awsConfig)

	for _, lmbd := range lambdaFunctions {
		lmbd.Role = iamLabRole.Arn
		opOut, err := lambdaSvc.CreateFunction(
			context.TODO(),
			&lmbd)
		if err != nil {
			log.Fatalf("unable to create function: %v", err)
		}

		log.Printf(
			"function %s, arn %s, status %s (reason %s)\n",
			*opOut.FunctionName,
			*opOut.FunctionArn,
			opOut.State,
			*opOut.StateReason)
	}
}

func createStepFunction() *string {
	sfnSvc := sfn.NewFromConfig(awsConfig)

	stateMachine.RoleArn = iamLabRole.Arn
	opOut, err := sfnSvc.CreateStateMachine(
		context.TODO(),
		&stateMachine)
	if err != nil {
		log.Fatalf("unable to create step function: %v", err)
	}

	log.Printf(
		"sfn arn %s\n",
		*opOut.StateMachineArn)

	return opOut.StateMachineArn
}

func createApiEndpoint() *string {
	apiSvc := apigatewayv2.NewFromConfig(awsConfig)

	opOut, err := apiSvc.CreateApi(
		context.TODO(),
		&api)
	if err != nil {
		log.Fatalf("unable to create API: %v", err)
	}

	log.Printf("api %s, endpoint %s, id %s\n ",
		*opOut.Name, *opOut.ApiEndpoint, *opOut.ApiId)

	return opOut.ApiId
}

func mergeRouteWithIntegration(
	apiId *string,
	integArn *string,
	integration apigatewayv2.CreateIntegrationInput,
	route apigatewayv2.CreateRouteInput) {

	apiSvc := apigatewayv2.NewFromConfig(awsConfig)

	integration.ApiId = apiId
	integration.CredentialsArn = iamLabRole.Arn
	integration.IntegrationUri = integArn

	integOpOut, err := apiSvc.CreateIntegration(
		context.TODO(),
		&integration)
	if err != nil {
		log.Fatalf("unable to create integration: %v", err)
	}

	log.Printf("integration %s\n", *integOpOut.IntegrationId)

	route.ApiId = apiId
	route.Target = integOpOut.IntegrationId

	routeOpOut, err := apiSvc.CreateRoute(
		context.TODO(),
		&route)
	if err != nil {
		log.Fatalf("unable to create route: %v", err)
	}

	log.Printf("route %s\n", *routeOpOut.RouteKey)
}

func obtainIamLabRole() {
	iamSvc := iam.NewFromConfig(awsConfig)

	roleInput := iam.GetRoleInput{RoleName: &IAM_LABROLE}
	ans, err := iamSvc.GetRole(
		context.TODO(),
		&roleInput)

	if err != nil {
		log.Fatalf("unable to retrieve info about role %s: %v",
			*roleInput.RoleName, err)
	}

	iamLabRole = *ans.Role
}

func deleteDynamoDbs() {

}

func deleteLambdas() {

}

func deleteStepFunction() {

}

func deleteRoutes() {

}

func deleteIntegrations() {

}

func deleteApiEndpoint() {

}

func main() {
	deleteAll := flag.Bool("--delete-all", false, "Delete all resources")

	awsCfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion(AWS_REGION))
	if err != nil {
		log.Fatalf("unable to load SDK config: %v", err)
	}

	awsConfig = awsCfg

	if !*deleteAll {
		obtainIamLabRole()

		createDynamoDbs()
		createLambdas()
		sfnArn := createStepFunction()
		apiId := createApiEndpoint()
		mergeRouteWithIntegration(apiId, sfnArn, integrations[0], routes[0])
	} else {
		deleteDynamoDbs()
		deleteLambdas()
		deleteStepFunction()
		deleteRoutes()
		deleteIntegrations()
		deleteApiEndpoint()
	}
}
