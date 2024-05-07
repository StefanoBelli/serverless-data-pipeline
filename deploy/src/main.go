package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	apigtypes "github.com/aws/aws-sdk-go-v2/service/apigatewayv2/types"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lmbdtypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	smtypes "github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"github.com/aws/aws-sdk-go-v2/service/sfn"
)

func dflCtx() context.Context {
	return context.TODO()
}

type AwsServiceClients struct {
	dynamodb       *dynamodb.Client
	apigateway     *apigatewayv2.Client
	sfn            *sfn.Client
	lambda         *lambda.Client
	iam            *iam.Client
	secretsmanager *secretsmanager.Client
}

var svc AwsServiceClients
var iamLabRoleArn string
var lambdasArns []string

/*
 * AWS create resources
 */

func createTables() {
	for _, table := range tables {
		opOut, err := svc.dynamodb.CreateTable(dflCtx(), &table)
		if err != nil {
			log.Printf("unable to create dynamodb table: %v\n", err)
		} else {
			log.Printf("create table %s, arn %s, status %s\n",
				*opOut.TableDescription.TableName,
				*opOut.TableDescription.TableArn,
				opOut.TableDescription.TableStatus)
		}
	}
}

func createApi() *string {
	_, err := getApiId()
	if err == nil {
		log.Printf("an api with this name: %s already exists\n", *api.Name)
		log.Println("auto-skipping route and integration creation")
		return nil
	}

	opOut, err := svc.apigateway.CreateApi(dflCtx(), &api)
	if err != nil {
		log.Printf("unable to create api: %v\n", err)
		return nil
	} else {
		log.Printf("create api %s, endpoint %s, id %s\n",
			*opOut.Name, *opOut.ApiEndpoint, *opOut.ApiId)
		return opOut.ApiId
	}
}

func createStepFunction() *string {
	lsmi := sfn.ListStateMachinesInput{MaxResults: 1000}

	for {
		lsmOut, err := svc.sfn.ListStateMachines(dflCtx(), &lsmi)
		if err != nil {
			log.Printf("unable to list state machines: %v\n", err)
			break
		}

		for _, smItem := range lsmOut.StateMachines {
			if *smItem.Name == *stateMachine.Name {
				log.Printf("unable to create sfn %s: already exists\n", *smItem.Name)
				return smItem.StateMachineArn
			}
		}

		lsmi.NextToken = lsmOut.NextToken
		if lsmi.NextToken == nil {
			break
		}
	}

	amlDef := getStateMachineDefinition()
	stateMachine.Definition = &amlDef

	opOut, err := svc.sfn.CreateStateMachine(dflCtx(), &stateMachine)
	if err != nil {
		log.Printf("unable to create step function: %v\n", err)
		return nil
	} else {
		log.Printf("create sfn arn %s\n", *opOut.StateMachineArn)
		return opOut.StateMachineArn
	}
}

func createLambdas(baseDir string) {
	for _, lmbd := range lambdas {
		zip, err := loadFunctionZip(baseDir, *lmbd.FunctionName)
		if err != nil {
			log.Printf("unable to load function zip: %v\n", err)
		} else {
			lmbd.Code = &lmbdtypes.FunctionCode{ZipFile: zip}
			opOut, err := svc.lambda.CreateFunction(dflCtx(), &lmbd)
			if err != nil {
				log.Printf("unable to create lambda %s: %v\n",
					*lmbd.FunctionName, err)
			} else {
				lambdasArns = append(lambdasArns, *opOut.FunctionArn)

				log.Printf("create lambda %s, arn: %s, state: %s (reason: %s)\n",
					*opOut.FunctionName, *opOut.FunctionArn,
					opOut.State, *opOut.StateReason)

				log.Printf("\twith deployment package of size %d B, sha256: %s, handler: %s\n",
					opOut.CodeSize, *opOut.CodeSha256,
					*opOut.Handler)
			}
		}
	}
}

func mergeRouteWithIntegration(apiId *string, sfnArn *string) *string {
	integration.ApiId = apiId
	integration.RequestParameters = make(map[string]string)
	integration.RequestParameters["StateMachineArn"] = *sfnArn

	integOpOut, err := svc.apigateway.CreateIntegration(dflCtx(), &integration)
	if err != nil {
		log.Printf("unable to create integration: %v\n", err)
		return nil
	} else {
		log.Printf("create integration %s, conn. type: %s, int. type: %s\n",
			*integOpOut.IntegrationId, integOpOut.ConnectionType,
			integOpOut.IntegrationType)

		route.ApiId = apiId
		myTarget := "integrations/" + *integOpOut.IntegrationId
		route.Target = &myTarget

		routeOpOut, err := svc.apigateway.CreateRoute(dflCtx(), &route)
		if err != nil {
			log.Printf("unable to create route: %v\n", err)
		} else {
			log.Printf("create route %s, id: %s, target: %s\n",
				*routeOpOut.RouteKey, *routeOpOut.RouteId, *routeOpOut.Target)
		}

		return routeOpOut.RouteId
	}
}

func addAuthorizerLambda() {
	lambdas = append(lambdas, lambda.CreateFunctionInput{
		FunctionName:  &authorizerS,
		Role:          &iamLabRoleArn,
		PackageType:   lmbdtypes.PackageTypeZip,
		Architectures: []lmbdtypes.Architecture{lmbdtypes.ArchitectureX8664},
		Runtime:       lmbdtypes.RuntimeProvidedal2023,
		Handler:       &bootstrap,
		Timeout:       &lambdaTimeout,
	})
}

func createOrUpdateSecret(key string) {
	secret.SecretBinary = []byte(key)
	csOut, err := svc.secretsmanager.CreateSecret(dflCtx(), &secret)
	if err != nil {
		psvi := secretsmanager.PutSecretValueInput{
			SecretId:     secret.Name,
			SecretBinary: secret.SecretBinary,
		}

		psvOut, err := svc.secretsmanager.PutSecretValue(dflCtx(), &psvi)
		if err != nil {
			log.Printf("unable to create or update secret: %v\n", err)
		} else {
			log.Printf("update secret (key [not shown]) for %s, arn: %s\n",
				*psvOut.Name, *psvOut.ARN)
		}
	} else {
		log.Printf("create secret %s (key [not shown]), arn: %s\n",
			*csOut.Name, *csOut.ARN)
	}
}

func createAuthorizer(apiId *string) string {
	authUri := getAuthorizerUri()
	authorizer.ApiId = apiId
	authorizer.AuthorizerUri = &authUri

	caOut, err := svc.apigateway.CreateAuthorizer(dflCtx(), &authorizer)
	if err != nil {
		log.Printf("unable to create authorizer: %v", err)

		return ""
	}

	log.Printf("create authorizer %s, id: %s, credArn: %s, "+
		"ttl: %d sec, type: %s, lambdaArn: %s, payld vers: %s\n",
		*caOut.Name, *caOut.AuthorizerId, *caOut.AuthorizerCredentialsArn,
		*caOut.AuthorizerResultTtlInSeconds, caOut.AuthorizerType, *caOut.AuthorizerUri,
		*caOut.AuthorizerPayloadFormatVersion)

	return *caOut.AuthorizerId
}

func addAuthorizerToRoute(authorizerId *string, routeId *string) {
	uri := apigatewayv2.UpdateRouteInput{
		ApiId:             route.ApiId,
		RouteId:           routeId,
		AuthorizationType: apigtypes.AuthorizationTypeCustom,
		AuthorizerId:      authorizerId,
	}

	urOut, err := svc.apigateway.UpdateRoute(dflCtx(), &uri)
	if err != nil {
		log.Printf("unable to update route %s: %v\n", *route.RouteKey, err)
	} else {
		log.Printf("update route %s (adding authorizer id: %s, with type: %s)\n",
			*urOut.RouteKey, *urOut.AuthorizerId, urOut.AuthorizationType)
	}
}

/*
 * AWS delete resources
 */

func deleteApi(apiId string) {
	dai := apigatewayv2.DeleteApiInput{ApiId: &apiId}
	_, err := svc.apigateway.DeleteApi(dflCtx(), &dai)
	if err != nil {
		log.Printf("unable to delete api: %v\n", err)
	} else {
		log.Printf("delete api id %s\n", *dai.ApiId)
	}
}

func deleteTables() {
	for _, table := range tables {
		dti := dynamodb.DeleteTableInput{TableName: table.TableName}
		opOut, err := svc.dynamodb.DeleteTable(dflCtx(), &dti)
		if err != nil {
			log.Printf("unable to delete table %s: %v\n", *dti.TableName, err)
		} else {
			log.Printf("delete table %s, arn: %s\n",
				*opOut.TableDescription.TableName,
				*opOut.TableDescription.TableArn)
		}
	}
}

func deleteLambdas() {
	addAuthorizerLambda()
	for _, lmbd := range lambdas {
		dfi := lambda.DeleteFunctionInput{FunctionName: lmbd.FunctionName}
		_, err := svc.lambda.DeleteFunction(dflCtx(), &dfi)
		if err != nil {
			log.Printf("unable to delete lambda %s: %v\n", *dfi.FunctionName, err)
		} else {
			log.Printf("delete lambda %s\n", *lmbd.FunctionName)
		}
	}
}

func deleteStepFunction() {
	lsmi := sfn.ListStateMachinesInput{MaxResults: 1000}

	for {
		lssmOut, err := svc.sfn.ListStateMachines(dflCtx(), &lsmi)
		if err != nil {
			log.Printf("unable to list state machines: %v\n", err)
			break
		} else {
			for _, sm := range lssmOut.StateMachines {
				if *sm.Name == *stateMachine.Name {
					dsmi := sfn.DeleteStateMachineInput{StateMachineArn: sm.StateMachineArn}
					_, err := svc.sfn.DeleteStateMachine(dflCtx(), &dsmi)
					if err != nil {
						log.Printf("unable to delete state machine %s: %v\n", *sm.Name, err)
					} else {
						log.Printf("delete sfn %s, arn: %s\n",
							*sm.Name, *sm.StateMachineArn)
					}

					return
				}
			}

			lsmi.NextToken = lssmOut.NextToken

			if lsmi.NextToken == nil {
				break
			}
		}
	}

	log.Printf("unable to find sfn %s\n", *stateMachine.Name)
}

func deleteRoutes(apiId string) {
	gri := apigatewayv2.GetRoutesInput{ApiId: &apiId, MaxResults: &s1000}

	for {
		grOut, err := svc.apigateway.GetRoutes(dflCtx(), &gri)
		if err != nil {
			log.Printf("unable to list routes: %v\n", err)
			break
		} else {
			for _, itemRoute := range grOut.Items {
				if *route.RouteKey == *itemRoute.RouteKey {
					dri := apigatewayv2.DeleteRouteInput{
						ApiId: &apiId, RouteId: itemRoute.RouteId}

					_, err := svc.apigateway.DeleteRoute(dflCtx(), &dri)
					if err != nil {
						log.Printf("unable to delete route: %v\n", err)
					} else {
						log.Printf("delete route %s (with api id: %s)\n",
							*dri.RouteId, *dri.ApiId)
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

func deleteIntegrations(apiId string) {
	gii := apigatewayv2.GetIntegrationsInput{ApiId: &apiId, MaxResults: &s1000}

	for {
		giOut, err := svc.apigateway.GetIntegrations(dflCtx(), &gii)
		if err != nil {
			log.Printf("unable to list integrations: %v\n", err)
			break
		} else {
			for _, itemIntegration := range giOut.Items {
				if *integration.Description == *itemIntegration.Description {
					dii := apigatewayv2.DeleteIntegrationInput{
						ApiId:         &apiId,
						IntegrationId: itemIntegration.IntegrationId,
					}
					_, err := svc.apigateway.DeleteIntegration(dflCtx(), &dii)
					if err != nil {
						log.Printf("unable to delete integration: %v\n", err)
					} else {
						log.Printf("delete integration %s (with api id: %s)\n",
							*dii.IntegrationId, *dii.ApiId)
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

func deleteSecret() {
	var two int32 = 2

	lsi := secretsmanager.ListSecretsInput{
		Filters: []smtypes.Filter{
			{
				Key:    smtypes.FilterNameStringTypeName,
				Values: []string{*authorizer.Name},
			},
		},
		MaxResults: &two,
		SortOrder:  smtypes.SortOrderTypeDesc,
	}

	lso, err := svc.secretsmanager.ListSecrets(dflCtx(), &lsi)
	if err != nil {
		log.Printf("unable to list secrets: %v\n", err)
	} else {
		var seven int64 = 7

		numSecrets := len(lso.SecretList)
		if numSecrets > 1 {
			log.Printf("WARNING unexpected number of results for secrets: %d (expected 1)\n", numSecrets)
		} else if numSecrets == 1 {
			dsi := secretsmanager.DeleteSecretInput{
				SecretId:                   lso.SecretList[0].ARN,
				ForceDeleteWithoutRecovery: &trueVal,
				RecoveryWindowInDays:       &seven,
			}
			dsOut, err := svc.secretsmanager.DeleteSecret(dflCtx(), &dsi)
			if err != nil {
				log.Printf("unable to delete secret: %v\n", err)
			} else {
				log.Printf("delete secret %s\n", *dsOut.Name)
			}
		}
	}
}

/*
 * AWS update resources
 */

func coreUpdateLambda(name *string, arch *[]lmbdtypes.Architecture, base *string) {
	zipBytes, err := loadFunctionZip(*base, *name)
	if err != nil {
		log.Printf("unable to load zip for %s: %v\n", *name, err)
	} else {
		ufci := lambda.UpdateFunctionCodeInput{
			FunctionName:  name,
			Architectures: *arch,
			ZipFile:       zipBytes,
		}

		opOut, err := svc.lambda.UpdateFunctionCode(dflCtx(), &ufci)
		if err != nil {
			log.Printf("unable to update lambda %s: %v\n",
				*ufci.FunctionName, err)
		} else {
			log.Printf("update lambda %s, arn: %s, state: %s\n",
				*opOut.FunctionName, *opOut.FunctionArn,
				opOut.State)

			log.Printf("\twith deployment package of size %d B, sha256: %s, handler: %s\n",
				opOut.CodeSize, *opOut.CodeSha256,
				*opOut.Handler)
		}
	}
}

func updateLambdas(base string, csl string) {
	if csl != "all" {
		lambdaNames := strings.Split(csl, ",")

		for _, lambdaName := range lambdaNames {
			lambdaName = strings.TrimSpace(lambdaName)
			if len(lambdaName) == 0 {
				continue
			}

			var archs []lmbdtypes.Architecture
			found := false

			for _, myLambda := range lambdas {
				if *myLambda.FunctionName == lambdaName {
					archs = myLambda.Architectures
					found = true
					break
				}
			}

			if !found && lambdaName != authorizerS {
				log.Printf("unable to find lambda %s\n", lambdaName)
				continue
			}

			coreUpdateLambda(&lambdaName, &archs, &base)
		}
	} else {
		for _, myLambda := range lambdas {
			coreUpdateLambda(myLambda.FunctionName, &myLambda.Architectures, &base)
		}
	}
}

/*
 * Various
 */

func getAuthorizerUri() string {
	funArn := &lambdasArns[len(lambdasArns)-1]
	return fmt.Sprintf(
		"arn:aws:apigateway:%s:lambda:path/2015-03-31/functions/%s/invocations",
		AWS_REGION,
		*funArn)
}

func obtainIamLabRole() {
	roleInput := iam.GetRoleInput{RoleName: &IAM_LABROLE}
	ans, err := svc.iam.GetRole(dflCtx(), &roleInput)
	if err != nil {
		log.Fatalf("unable to retrieve info about role %s: %v",
			*roleInput.RoleName, err)
	}

	iamLabRoleArn = *ans.Role.Arn
}

func getApiId() (string, error) {
	gasi := apigatewayv2.GetApisInput{MaxResults: &s1000}

	for {
		gaso, err := svc.apigateway.GetApis(dflCtx(), &gasi)
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
			break
		}
	}

	return "", errors.New("unable to find api")
}

func getStateMachineDefinition() string {
	return fmt.Sprintf(SFN_AML_DEFINITION_FMT,
		lambdas[0].FunctionName, //validate
		lambdas[1].FunctionName, //transform
		lambdas[4].FunctionName, //flagTransformFailed
		lambdas[3].FunctionName, //flagValidateFailed
		lambdas[2].FunctionName, //store
		lambdas[5].FunctionName, //flagStoreFailed
		lambdas[4].FunctionName, //flagTransformFailed
		lambdas[3].FunctionName, //flagValidateFailed
		lambdas[3].FunctionName, //flagValidateFailed
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
				path.Clean(awsCredentialsFile))
		}
	}
}

type Cmdline struct {
	baseLambdaPkgs   string
	deleteAll        bool
	updateLambdas    string
	authorizationKey string
	forceSecretDel   bool
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

	flag.StringVar(
		&cmdline.updateLambdas,
		"u",
		"",
		"Comma-separated lambdas to update or all",
	)

	flag.StringVar(
		&cmdline.authorizationKey,
		"a",
		"",
		"Enable authorization. Key is to be entered "+
			"on \"Authorization\" http header when making requests",
	)

	flag.BoolVar(
		&cmdline.forceSecretDel,
		"s",
		false,
		"Force secret deletion along with all the other resources."+
			" You will not be able to create another secret with same name"+
			" for the next 7 days.",
	)

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

	svc.dynamodb = dynamodb.NewFromConfig(awsCfg)
	svc.lambda = lambda.NewFromConfig(awsCfg)
	svc.sfn = sfn.NewFromConfig(awsCfg)
	svc.apigateway = apigatewayv2.NewFromConfig(awsCfg)
	svc.iam = iam.NewFromConfig(awsCfg)
	svc.secretsmanager = secretsmanager.NewFromConfig(awsCfg)
}

func beginIgnoreInterruption() chan os.Signal {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			//consume
		}
	}()

	return c
}

func endIgnoreInteruption(c chan os.Signal) {
	signal.Reset(os.Interrupt)
	close(c)
}

func recoverIfNeeded(apiId **string) {
	if *apiId == nil {
		recoveredApiId, err := getApiId()
		if err != nil {
			log.Fatalf("unable to proceed, no api id: %v", err)
		}

		*apiId = new(string)
		**apiId = recoveredApiId
	}

	if len(lambdasArns) == 0 {
		gfi := lambda.GetFunctionInput{
			FunctionName: lambdas[len(lambdas)-1].FunctionName,
		}
		gfOut, err := svc.lambda.GetFunction(dflCtx(), &gfi)
		if err == nil {
			lambdasArns = append(lambdasArns, *gfOut.Configuration.FunctionArn)
		} else {
			log.Fatalf("unable to proceed, no authorizer-lambda arn: %v", err)
		}
	}
}

func main() {
	checkAwsCredentialsFile()

	cmdline := parseCmdline()

	loadAwsConfig()

	if len(cmdline.updateLambdas) > 0 {
		updateLambdas(cmdline.baseLambdaPkgs, cmdline.updateLambdas)
	} else {
		if !cmdline.deleteAll {
			authRequired := len(cmdline.authorizationKey) > 0

			obtainIamLabRole()

			createTables()
			if authRequired {
				addAuthorizerLambda()
			}
			createLambdas(cmdline.baseLambdaPkgs)
			sfnArn := createStepFunction()
			intChan := beginIgnoreInterruption()
			apiId := createApi()
			routeId := mergeRouteWithIntegration(apiId, sfnArn)
			endIgnoreInteruption(intChan)

			if authRequired {
				createOrUpdateSecret(cmdline.authorizationKey)
				recoverIfNeeded(&apiId)

				authorizerId := createAuthorizer(apiId)

				if authorizerId != "" {
					addAuthorizerToRoute(&authorizerId, routeId)
				}
			}
		} else {
			deleteTables()
			deleteLambdas()
			deleteStepFunction()
			intChan := beginIgnoreInterruption()
			apiId, err := getApiId()
			if err != nil {
				log.Println("unable to get api id")
				log.Println("cannot go further")
				log.Fatalf("stopping immediately (reason: %v)", err)
			}
			deleteRoutes(apiId)
			deleteIntegrations(apiId)
			deleteApi(apiId)
			endIgnoreInteruption(intChan)

			if cmdline.forceSecretDel {
				deleteSecret() //try deletion anyway
			} else {
				log.Println("skipping secret deletion")
			}
		}
	}
}
