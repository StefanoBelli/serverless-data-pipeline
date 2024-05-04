package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
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

	amlDef := getStateMachineDefinition()
	stateMachine.Definition = &amlDef

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

func mergeRouteWithIntegration(apiId *string, integArn *string,
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
	dynamoDbSvc := dynamodb.NewFromConfig(awsConfig)

	for _, ddbTable := range dynamoDbTables {
		dti := dynamodb.DeleteTableInput{TableName: ddbTable.TableName}
		opOut, err := dynamoDbSvc.DeleteTable(context.TODO(), &dti)
		if err != nil {
			log.Printf("unable to delete table %s: %v\n",
				*opOut.TableDescription.TableName, err)
		} else {
			log.Printf("delete table %s\n",
				*opOut.TableDescription.TableName)
		}
	}
}

func deleteLambdas() {
	lambdaSvc := lambda.NewFromConfig(awsConfig)

	for _, lmbd := range lambdaFunctions {
		dfi := lambda.DeleteFunctionInput{FunctionName: lmbd.FunctionName}
		_, err := lambdaSvc.DeleteFunction(context.TODO(), &dfi)
		if err != nil {
			log.Printf("unable to delete lambda %s: %v\n",
				*dfi.FunctionName, err)
		} else {
			log.Printf("delete lambda %s\n",
				*dfi.FunctionName)
		}
	}
}

type StateMachineInfo struct {
	name string
	arn  string
}

func deleteStepFunction() {
	sfnSvc := sfn.NewFromConfig(awsConfig)

	lsmi := sfn.ListStateMachinesInput{MaxResults: 1000}

	var smInfo []StateMachineInfo
	for {
		lsmo, err := sfnSvc.ListStateMachines(context.TODO(), &lsmi)
		if err != nil {
			log.Fatalf("unable to list state machines: %v", err)
		}

		for _, mySm := range lsmo.StateMachines {
			smInfo = append(
				smInfo,
				StateMachineInfo{name: *mySm.Name, arn: *mySm.StateMachineArn})
		}

		lsmi.NextToken = lsmo.NextToken

		if lsmi.NextToken == nil {
			break
		}
	}

	for _, sm := range smInfo {
		if sm.name == *stateMachine.Name {
			dsmi := sfn.DeleteStateMachineInput{StateMachineArn: &sm.arn}
			_, err := sfnSvc.DeleteStateMachine(context.TODO(), &dsmi)
			if err != nil {
				log.Printf("unable to delete state machine %s: %v\n",
					sm.name, err)
			} else {
				log.Printf("delete state machine %s\n", sm.name)
			}

			break
		}
	}
}

func getApiId() (string, error) {
	apiSvc := apigatewayv2.NewFromConfig(awsConfig)

	s1000 := "1000"
	gasi := apigatewayv2.GetApisInput{MaxResults: &s1000}

	for {
		gaso, err := apiSvc.GetApis(context.TODO(), &gasi)
		if err != nil {
			log.Fatalf("unable to list apis: %v", err)
		}

		for _, apiItem := range gaso.Items {
			if api.Name == apiItem.Name {
				return *apiItem.ApiId, nil
			}
		}

		gasi.NextToken = gaso.NextToken
		if gasi.NextToken == nil {
			return "", errors.New("unable to find api")
		}
	}
}

func deleteRoutes(apiId string) {
	apiSvc := apigatewayv2.NewFromConfig(awsConfig)

	s1000 := "1000"
	gri := apigatewayv2.GetRoutesInput{ApiId: &apiId, MaxResults: &s1000}

	for {
		gro, err := apiSvc.GetRoutes(context.TODO(), &gri)
		if err != nil {
			log.Fatalf("unable to list routes: %v", err)
		}

		for _, rItem := range gro.Items {
			for _, rHc := range routes {
				if *rHc.RouteKey == *rItem.RouteKey {
					dri := apigatewayv2.DeleteRouteInput{ApiId: &apiId, RouteId: rItem.RouteId}
					_, err := apiSvc.DeleteRoute(context.TODO(), &dri)
					if err != nil {
						log.Printf("unable to delete route: %v\n", err)
					} else {
						log.Printf("delete route %s\n", *rHc.RouteKey)
					}
				}
			}
		}

		gri.NextToken = gro.NextToken
		if gri.NextToken == nil {
			break
		}
	}
}

func deleteIntegration(apiId string) {
	apiSvc := apigatewayv2.NewFromConfig(awsConfig)

	s1000 := "1000"
	gii := apigatewayv2.GetIntegrationsInput{ApiId: &apiId, MaxResults: &s1000}

	for {
		gio, err := apiSvc.GetIntegrations(context.TODO(), &gii)
		if err != nil {
			log.Fatalf("unable to list integrations: %v", err)
		}

		for _, gItem := range gio.Items {
			for _, iHc := range integrations {
				if *gItem.Description == *iHc.Description {
					dii := apigatewayv2.DeleteIntegrationInput{ApiId: &apiId, IntegrationId: gItem.IntegrationId}
					_, err := apiSvc.DeleteIntegration(context.TODO(), &dii)
					if err != nil {
						log.Printf("unable to delete integration: %v\n", err)
					} else {
						log.Printf("delete integration %s\n", *gItem.IntegrationId)
					}
				}
			}
		}

		gii.NextToken = gio.NextToken
		if gii.NextToken == nil {
			break
		}
	}
}

func getStateMachineDefinition() string {
	return fmt.Sprintf(SFN_AML_DEFINITION_FMT,
		lambdaFunctions[0].FunctionName, //validate
		lambdaFunctions[1].FunctionName, //transform
		lambdaFunctions[4].FunctionName, //flagTransformFailed
		lambdaFunctions[3].FunctionName, //flagValidateFailed
		lambdaFunctions[2].FunctionName, //store
		lambdaFunctions[5].FunctionName, //flagStoreFailed
		lambdaFunctions[4].FunctionName, //flagTransformFailed
		lambdaFunctions[3].FunctionName, //flagValidateFailed
		lambdaFunctions[3].FunctionName, //flagValidateFailed
	)
}

func main() {
	deleteAll := flag.Bool("d", false, "Delete all resources")

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
		apiId, err := getApiId()
		if err != nil {
			log.Fatalf("unable to get API id")
		}
		deleteRoutes(apiId)
		deleteIntegration(apiId)
	}
}
