package database

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func GetDatabase(ctx context.Context, awsRegion, awsEndpoint string) (*dynamodb.Client, error) {
	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(awsRegion),
	)
	if err != nil {
		return &dynamodb.Client{}, fmt.Errorf("cannot load the AWS configs: %w", err)
	}

	var opts []func(*dynamodb.Options)
	if awsEndpoint != "" {
		opts = append(opts, func(o *dynamodb.Options) {
			o.BaseEndpoint = aws.String(awsEndpoint)
		})
	}

	client := dynamodb.NewFromConfig(awsCfg, opts...)

	return client, nil
}

func GetS3Client(ctx context.Context, awsRegion, awsEndpoint string) (*s3.Client, error) {
	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(awsRegion),
	)
	if err != nil {
		return &s3.Client{}, fmt.Errorf("cannot load the AWS configs: %w", err)
	}

	opts := []func(o *s3.Options){func(o *s3.Options) {
		o.UsePathStyle = true
	}}

	if awsEndpoint != "" {
		opts = append(opts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(awsEndpoint)
		})
	}

	client := s3.NewFromConfig(awsCfg, opts...)

	return client, nil
}
