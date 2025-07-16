package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/HENNGE/snsclone-202506-golang-luca/database"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/joho/godotenv"
)

func InitTable(ctx context.Context, client *dynamodb.Client, tableName string) error {
	t, err := client.DescribeTable(
		ctx, &dynamodb.DescribeTableInput{TableName: aws.String(tableName)},
	)
	if err == nil {
		// Cond: No error = Table exists nothing to do so return early
		log.Printf("Table %s already exists, skipping initialization\n", *t.Table.TableName)
		return nil
	}

	// There was an error handle it based on the type
	var notFoundEx *types.ResourceNotFoundException
	if !errors.As(err, &notFoundEx) {
		// cond: Not NotFoundException = unexpected result, return the error
		return fmt.Errorf("failed to describe table %s: %w", tableName, err)
	}

	log.Printf("Table %s does not exist, creating it...\n", tableName)

	attributeDefinitions := []types.AttributeDefinition{
		{
			AttributeName: aws.String("pk"),
			AttributeType: types.ScalarAttributeTypeS,
		},
		{
			AttributeName: aws.String("sk"),
			AttributeType: types.ScalarAttributeTypeS,
		},
		{
			AttributeName: aws.String("gsi1_pk"),
			AttributeType: types.ScalarAttributeTypeS,
		},
		{
			AttributeName: aws.String("gsi1_sk"),
			AttributeType: types.ScalarAttributeTypeS,
		},
	}

	keySchema := []types.KeySchemaElement{
		{
			AttributeName: aws.String("pk"),
			KeyType:       types.KeyTypeHash,
		},
		{
			AttributeName: aws.String("sk"),
			KeyType:       types.KeyTypeRange,
		},
	}

	gsi := []types.GlobalSecondaryIndex{
		{
			IndexName: aws.String("gsi1"),
			KeySchema: []types.KeySchemaElement{
				{
					AttributeName: aws.String("gsi1_pk"),
					KeyType:       types.KeyTypeHash,
				},
				{
					AttributeName: aws.String("gsi1_sk"),
					KeyType:       types.KeyTypeRange,
				},
			},
			Projection: &types.Projection{
				ProjectionType: types.ProjectionTypeAll,
			},
		},
	}

	_, err = client.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName:              &tableName,
		KeySchema:              keySchema,
		AttributeDefinitions:   attributeDefinitions,
		GlobalSecondaryIndexes: gsi,
		BillingMode:            types.BillingModePayPerRequest,
	})
	if err != nil {
		return fmt.Errorf("failed to create table %s: %w", tableName, err)
	}

	log.Printf("Table %s created successfully...\n", tableName)

	return nil
}

func main() {
	ctx := context.Background()

	if err := godotenv.Load(); err != nil {
		log.Fatal("No .env file found")
	}

	awsRegion, exists := os.LookupEnv("AWS_REGION")
	if !exists {
		log.Fatal("Undefined AWS region")
	}

	awsEndpoint, exists := os.LookupEnv("AWS_ENDPOINT")
	if !exists {
		log.Fatal("Undefined AWS endpoint")
	}

	db, err := database.GetDatabase(ctx, awsRegion, awsEndpoint)
	if err != nil {
		log.Fatal("failed to get database: ", err)
	}

	tableName, exists := os.LookupEnv("TABLE_NAME")
	if !exists {
		log.Fatal("Undefined table name")
	}

	err = InitTable(ctx, db, tableName)
	if err != nil {
		log.Fatal("Failed to initialize table: ", err)
	}
}
