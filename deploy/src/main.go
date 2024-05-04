package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/sfn"
)

func dflCtx() context.Context {
	return context.TODO()
}

var awsConfig aws.Config
var iamLabRole iamtypes.Role

/*
 * AWS deploy
 */

func createDynamoDbs() {
	dynamoDbSvc := dynamodb.NewFromConfig(awsConfig)

	for _, dbs := range dynamoDbTables {
		opOut, err := dynamoDbSvc.CreateTable(dflCtx(), &dbs)
		if err != nil {
			log.Printf("unable to create dynamodb table: %v\n", err)
		} else {
			log.Printf
		}
	}
}

func createApiEndpoint() *string {
	apiSvc := apigatewayv2.NewFromConfig(awsConfig)

	opOut, err := apiSvc.CreateApi(dflCtx(), &api)
	if err != nil {
		log.Printf("unable to create API: %v\n", err)
		return nil
	} else {
		log.Printf
		return opOut.ApiId
	}
}

func createStepFunction() *string {
	sfnSvc := sfn.NewFromConfig(awsConfig)

	amlDef := getStateMachineDefinition()
	stateMachine.Definition = &amlDef
	stateMachine.RoleArn = iamLabRole.Arn

	opOut, err := sfnSvc.CreateStateMachine(dflCtx(), &stateMachine)
	if err != nil {
		log.Printf("unable to create step function: %v\n", err)
		return nil
	} else {
		log.Printf
		return opOut.StateMachineArn
	}
}

func createAppDomainLambdas(baseDir string) {
	lambdaSvc := lambda.NewFromConfig(awsConfig)

	for _, lmbd := range appDomainLambdas {
		zip, err := loadFunctionZip(baseDir, *lmbd.FunctionName)
		if err != nil {
			log.Printf("unable to load function zip: %v\n", err)
		} else {
			lmbd.Code.ZipFile = zip
			lmbd.Role = iamLabRole.Arn
			opOut, err := lambdaSvc.CreateFunction(dflCtx(), &lmbd)
			if err != nil {
				log.Printf("unable to create lambda: %v\n", err)
			} else {
				log.Printf
			}
		}
	}
}

type MergeRouteIntegration struct {
	apiId           *string
	arnParameterKey string
	integArn        *string
	integration     apigatewayv2.CreateIntegrationInput
	route           apigatewayv2.CreateRouteInput
}

func mergeRouteWithIntegration(merge *MergeRouteIntegration) {

	apiSvc := apigatewayv2.NewFromConfig(awsConfig)

	merge.integration.ApiId = merge.apiId
	merge.integration.CredentialsArn = iamLabRole.Arn
	merge.integration.RequestParameters = make(map[string]string)
	merge.integration.RequestParameters[merge.arnParameterKey] = *merge.integArn

	integOpOut, err := apiSvc.CreateIntegration(dflCtx(), &merge.integration)
	if err != nil {
		log.Printf("unable to create integration: %v\n", err)
	} else {
		log.Printf("integration %s\n", *integOpOut.IntegrationId)

		merge.route.ApiId = merge.apiId
		myTarget := "integrations/" + *integOpOut.IntegrationId
		merge.route.Target = &myTarget

		routeOpOut, err := apiSvc.CreateRoute(
			dflCtx(),
			&merge.route)
		if err != nil {
			log.Printf("unable to create route: %v\n", err)
		} else {
			log.Printf
		}
	}
}

/*
 * AWS undeploy
 */

func deleteApi(apiId string) {
	apiSvc := apigatewayv2.NewFromConfig(awsConfig)

	dai := apigatewayv2.DeleteApiInput{ApiId: &apiId}
	_, err := apiSvc.DeleteApi(dflCtx(), &dai)
	if err != nil {
		log.Printf("unable to delete API: %v\n", err)
	} else {
		log.Printf
	}
}

func deleteDynamoDbs() {
	dynamoDbSvc := dynamodb.NewFromConfig(awsConfig)

	for _, ddbTable := range dynamoDbTables {
		dti := dynamodb.DeleteTableInput{TableName: ddbTable.TableName}
		opOut, err := dynamoDbSvc.DeleteTable(dflCtx(), &dti)
		if err != nil {
			log.Printf("unable to delete table %s: %v\n", *dti.TableName, err)
		} else {
			log.Printf
		}
	}
}

func deleteLambdas() {
	lambdaSvc := lambda.NewFromConfig(awsConfig)

	for _, lmbd := range appDomainLambdas {
		dfi := lambda.DeleteFunctionInput{FunctionName: lmbd.FunctionName}
		_, err := lambdaSvc.DeleteFunction(dflCtx(), &dfi)
		if err != nil {
			log.Printf("unable to delete lambda %s: %v\n", *dfi.FunctionName, err)
		} else {
			log.Printf
		}
	}
}

func deleteStepFunction() {
	sfnSvc := sfn.NewFromConfig(awsConfig)

	lsmi := sfn.ListStateMachinesInput{MaxResults: 1000}

	for {
		lssmOut, err := sfnSvc.ListStateMachines(dflCtx(), &lsmi)
		if err != nil {
			log.Printf("unable to list state machines: %v\n", err)
			break
		} else {
			for _, sm := range lssmOut.StateMachines {
				if *sm.Name == *stateMachine.Name {
					dsmi := sfn.DeleteStateMachineInput{StateMachineArn: sm.StateMachineArn}
					_, err := sfnSvc.DeleteStateMachine(dflCtx(), &dsmi)
					if err != nil {
						log.Printf("unable to delete state machine %s: %v\n", *sm.Name, err)
					} else {
						log.Printf
					}

					break
				}
			}

			lsmi.NextToken = lssmOut.NextToken

			if lsmi.NextToken == nil {
				break
			}
		}
	}
}

func deleteRoutes(apiId string) {
	apiSvc := apigatewayv2.NewFromConfig(awsConfig)

	gri := apigatewayv2.GetRoutesInput{ApiId: &apiId, MaxResults: &s1000}

	for {
		grOut, err := apiSvc.GetRoutes(dflCtx(), &gri)
		if err != nil {
			log.Printf("unable to list routes: %v\n", err)
			break
		} else {
			for _, myRoute := range routes {
				for _, route := range grOut.Items {
					if *myRoute.RouteKey == *route.RouteKey {
						dri := apigatewayv2.DeleteRouteInput{
							ApiId: &apiId, RouteId: route.RouteId}

						_, err := apiSvc.DeleteRoute(dflCtx(), &dri)
						if err != nil {
							log.Printf("unable to delete route: %v\n", err)
						} else {
							log.Printf
						}
					}
				}
			}

			gri.NextToken = grOut.NextToken
			if gri.NextToken == nil {
				break
			}
		}
	}
}

func deleteIntegration(apiId string) {
	apiSvc := apigatewayv2.NewFromConfig(awsConfig)

	gii := apigatewayv2.GetIntegrationsInput{ApiId: &apiId, MaxResults: &s1000}

	for {
		giOut, err := apiSvc.GetIntegrations(dflCtx(), &gii)
		if err != nil {
			log.Printf("unable to list integrations: %v\n", err)
			break
		} else {
			for _, myIntegration := range integrations {
				for _, integration := range giOut.Items {
					if *integration.Description == *myIntegration.Description {
						dii := apigatewayv2.DeleteIntegrationInput{
							ApiId: &apiId, IntegrationId: integration.IntegrationId}

						_, err := apiSvc.DeleteIntegration(dflCtx(), &dii)
						if err != nil {
							log.Printf("unable to delete integration: %v\n", err)
						} else {
							log.Printf
						}
					}
				}
			}

			gii.NextToken = giOut.NextToken
			if gii.NextToken == nil {
				break
			}
		}
	}
}

/*
 * AWS util
 */

func obtainIamLabRole() {
	iamSvc := iam.NewFromConfig(awsConfig)

	roleInput := iam.GetRoleInput{RoleName: &IAM_LABROLE}
	ans, err := iamSvc.GetRole(
		dflCtx(),
		&roleInput)

	if err != nil {
		log.Fatalf("unable to retrieve info about role %s: %v",
			*roleInput.RoleName, err)
	}

	iamLabRole = *ans.Role
}

func getApiId() (string, error) {
	apiSvc := apigatewayv2.NewFromConfig(awsConfig)

	gasi := apigatewayv2.GetApisInput{MaxResults: &s1000}

	for {
		gaso, err := apiSvc.GetApis(dflCtx(), &gasi)
		if err != nil {
			return "", err
		}

		for _, apiItem := range gaso.Items {
			if *api.Name == *apiItem.Name {
				return *apiItem.ApiId, nil
			}
		}

		gasi.NextToken = gaso.NextToken
		if gasi.NextToken == nil {
			return "", errors.New("unable to find api")
		}
	}
}

func getStateMachineDefinition() string {
	return fmt.Sprintf(SFN_AML_DEFINITION_FMT,
		appDomainLambdas[0].FunctionName, //validate
		appDomainLambdas[1].FunctionName, //transform
		appDomainLambdas[4].FunctionName, //flagTransformFailed
		appDomainLambdas[3].FunctionName, //flagValidateFailed
		appDomainLambdas[2].FunctionName, //store
		appDomainLambdas[5].FunctionName, //flagStoreFailed
		appDomainLambdas[4].FunctionName, //flagTransformFailed
		appDomainLambdas[3].FunctionName, //flagValidateFailed
		appDomainLambdas[3].FunctionName, //flagValidateFailed
	)
}

func loadFunctionZip(pkgs string, name string) ([]byte, error) {
	path := pkgs + "/" + name + "/" + name + ".zip"
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil, err
	}

	zipBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return zipBytes, nil
}

func checkAwsCredentialsFile() {
	if homeDir, err := os.UserHomeDir(); err != nil {
		log.Printf("unable to find home directory (%v)\n", err)
		log.Println("skipping aws credentials file existence check")
	} else {
		awsCredentialsFile := homeDir + "/.aws/credentials"
		_, err := os.Stat(awsCredentialsFile)
		if os.IsNotExist(err) {
			log.Fatalf("unable to find credentials file (looked: %s)",
				awsCredentialsFile)
		}
	}
}

type Cmdline struct {
	baseLambdaPkgs string
	deleteAll      bool
}

func parseCmdline() Cmdline {
	var cmdline Cmdline

	flag.StringVar(
		&cmdline.baseLambdaPkgs,
		"p",
		"../../lambdas/pkgs",
		"BaseDir for built lambda deployment packages")

	flag.BoolVar(
		&cmdline.deleteAll,
		"d",
		false,
		"Delete all resources")

	flag.Parse()

	return cmdline
}

func loadAwsConfig() {
	awsCfg, err := config.LoadDefaultConfig(
		dflCtx(),
		config.WithRegion(AWS_REGION))
	if err != nil {
		log.Fatalf("unable to load SDK config: %v", err)
	}

	awsConfig = awsCfg
}

func main() {
	checkAwsCredentialsFile()

	cmdline := parseCmdline()

	loadAwsConfig()

	if !cmdline.deleteAll {
		obtainIamLabRole()

		createDynamoDbs()
		createAppDomainLambdas(cmdline.baseLambdaPkgs)
		apiId := createApiEndpoint()
		sfnArn := createStepFunction()
		if sfnArn != nil && apiId != nil {
			mri := MergeRouteIntegration{
				apiId:           apiId,
				integArn:        sfnArn,
				arnParameterKey: "StateMachineArn",
				integration:     integrations[0],
				route:           routes[0],
			}
			mergeRouteWithIntegration(&mri)
		} else {
			log.Println("unable to merge route with integration")
		}
	} else {
		deleteDynamoDbs()
		deleteLambdas()
		deleteStepFunction()
		apiId, err := getApiId()
		if err != nil {
			log.Println("unable to get API id")
			log.Println("cannot go further")
			log.Fatalf("stopping immediately (reason: %v)", err)
		}
		deleteRoutes(apiId)
		deleteIntegration(apiId)
		deleteApi(apiId)
	}
}
