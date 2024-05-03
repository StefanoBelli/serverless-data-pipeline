package main

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lmbdtypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
)

const AWS_REGION = "us-east-1"

var dynamoDbTables = []dynamodb.CreateTableInput{
	{
		TableName: nil,
		AttributeDefinitions: []ddbtypes.AttributeDefinition{
			{},
			{},
		},
		KeySchema: []ddbtypes.KeySchemaElement{
			{},
		},
	},
	{
		TableName: nil,
		AttributeDefinitions: []ddbtypes.AttributeDefinition{
			{},
			{},
		},
		KeySchema: []ddbtypes.KeySchemaElement{
			{},
		},
	},
	{
		TableName: nil,
		AttributeDefinitions: []ddbtypes.AttributeDefinition{
			{},
			{},
		},
		KeySchema: []ddbtypes.KeySchemaElement{
			{},
		},
	},
	{
		TableName: nil,
		AttributeDefinitions: []ddbtypes.AttributeDefinition{
			{},
			{},
		},
		KeySchema: []ddbtypes.KeySchemaElement{
			{},
		},
	},
	{
		TableName: nil,
		AttributeDefinitions: []ddbtypes.AttributeDefinition{
			{},
			{},
		},
		KeySchema: []ddbtypes.KeySchemaElement{
			{},
		},
	},
	{
		TableName: nil,
		AttributeDefinitions: []ddbtypes.AttributeDefinition{
			{},
			{},
		},
		KeySchema: []ddbtypes.KeySchemaElement{
			{},
		},
	},
}

var lambdaFunctions = []lambda.CreateFunctionInput{
	{
		FunctionName:  nil,
		Description:   nil,
		Code:          &lmbdtypes.FunctionCode{ZipFile: nil},
		PackageType:   lmbdtypes.PackageTypeZip,
		Role:          nil,
		Architectures: []lmbdtypes.Architecture{lmbdtypes.ArchitectureX8664},
		Runtime:       lmbdtypes.RuntimeGo1x,
		Timeout:       nil,
	},
	{
		FunctionName:  nil,
		Description:   nil,
		Code:          &lmbdtypes.FunctionCode{ZipFile: nil},
		PackageType:   lmbdtypes.PackageTypeZip,
		Role:          nil,
		Architectures: []lmbdtypes.Architecture{lmbdtypes.ArchitectureX8664},
		Runtime:       lmbdtypes.RuntimeGo1x,
		Timeout:       nil,
	},
	{
		FunctionName:  nil,
		Description:   nil,
		Code:          &lmbdtypes.FunctionCode{ZipFile: nil},
		PackageType:   lmbdtypes.PackageTypeZip,
		Role:          nil,
		Architectures: []lmbdtypes.Architecture{lmbdtypes.ArchitectureX8664},
		Runtime:       lmbdtypes.RuntimeGo1x,
		Timeout:       nil,
	},
	{
		FunctionName:  nil,
		Description:   nil,
		Code:          &lmbdtypes.FunctionCode{ZipFile: nil},
		PackageType:   lmbdtypes.PackageTypeZip,
		Role:          nil,
		Architectures: []lmbdtypes.Architecture{lmbdtypes.ArchitectureX8664},
		Runtime:       lmbdtypes.RuntimeGo1x,
		Timeout:       nil,
	},
	{
		FunctionName:  nil,
		Description:   nil,
		Code:          &lmbdtypes.FunctionCode{ZipFile: nil},
		PackageType:   lmbdtypes.PackageTypeZip,
		Role:          nil,
		Architectures: []lmbdtypes.Architecture{lmbdtypes.ArchitectureX8664},
		Runtime:       lmbdtypes.RuntimeGo1x,
		Timeout:       nil,
	},
	{
		FunctionName:  nil,
		Description:   nil,
		Code:          &lmbdtypes.FunctionCode{ZipFile: nil},
		PackageType:   lmbdtypes.PackageTypeZip,
		Role:          nil,
		Architectures: []lmbdtypes.Architecture{lmbdtypes.ArchitectureX8664},
		Runtime:       lmbdtypes.RuntimeGo1x,
		Timeout:       nil,
	},
}
