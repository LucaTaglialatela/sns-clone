package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/HENNGE/snsclone-202506-golang-luca/database"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/joho/godotenv"
)

func InitBucket(ctx context.Context, client *s3.Client, baseUrl, bucketName string, region string) error {
	_, err := client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err == nil {
		log.Printf("Bucket %s already exists, skipping initialization\n", bucketName)
		return nil
	}

	var notFoundEx *types.NotFound
	if !errors.As(err, &notFoundEx) {
		return fmt.Errorf("failed to describe bucket %s: %w", bucketName, err)
	}

	log.Printf("Bucket %s does not exist, creating it...\n", bucketName)

	var locationConstraint types.BucketLocationConstraint
	if region != "us-east-1" {
		locationConstraint = types.BucketLocationConstraint(region)
	}

	_, err = client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: &bucketName,
		CreateBucketConfiguration: &types.CreateBucketConfiguration{
			LocationConstraint: locationConstraint,
		},
	})

	if err != nil {
		return fmt.Errorf("failed to create bucket %s: %w", bucketName, err)
	}

	log.Printf("Bucket %s created successfully...\n", bucketName)

	log.Printf("Applying CORS configuration to bucket %s...", bucketName)

	corsInput := &s3.PutBucketCorsInput{
		Bucket: aws.String(bucketName),
		CORSConfiguration: &types.CORSConfiguration{
			CORSRules: []types.CORSRule{
				{
					AllowedHeaders: []string{"*"},
					AllowedMethods: []string{"PUT", "POST", "GET"},
					AllowedOrigins: []string{baseUrl},
					ExposeHeaders:  []string{"ETag"},
				},
			},
		},
	}

	_, err = client.PutBucketCors(ctx, corsInput)
	if err != nil {
		return fmt.Errorf("failed to apply CORS configuration to bucket %s: %w", bucketName, err)
	}

	log.Println("CORS configuration applied successfully.")
	return nil
}

func DeleteBucket(ctx context.Context, client *s3.Client, bucketName string) error {
	log.Printf("Attempting to delete bucket: %s", bucketName)

	log.Println("Listing objects to delete...")
	listObjectsInput := &s3.ListObjectsV2Input{
		Bucket: &bucketName,
	}

	paginator := s3.NewListObjectsV2Paginator(client, listObjectsInput)

	var objectIdentifiers []types.ObjectIdentifier
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to list objects for bucket %s: %w", bucketName, err)
		}
		for _, obj := range page.Contents {
			objectIdentifiers = append(objectIdentifiers, types.ObjectIdentifier{Key: obj.Key})
		}
	}

	if len(objectIdentifiers) > 0 {
		log.Printf("Deleting %d objects...", len(objectIdentifiers))
		deleteObjectsInput := &s3.DeleteObjectsInput{
			Bucket: &bucketName,
			Delete: &types.Delete{Objects: objectIdentifiers},
		}
		_, err := client.DeleteObjects(ctx, deleteObjectsInput)
		if err != nil {
			return fmt.Errorf("failed to delete objects from bucket %s: %w", bucketName, err)
		}
		log.Println("All objects deleted successfully.")
	} else {
		log.Println("Bucket is already empty.")
	}

	log.Println("Deleting the bucket...")
	deleteBucketInput := &s3.DeleteBucketInput{
		Bucket: &bucketName,
	}
	_, err := client.DeleteBucket(ctx, deleteBucketInput)
	if err != nil {
		return fmt.Errorf("failed to delete bucket %s: %w", bucketName, err)
	}

	log.Printf("Bucket %s deleted successfully.", bucketName)
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

	s3Client, err := database.GetS3Client(ctx, awsRegion, awsEndpoint)
	if err != nil {
		log.Fatal("Failed to get s3 client")
	}

	bucketName, exists := os.LookupEnv("BUCKET_NAME")
	if !exists {
		log.Fatal("Undefined bucket name")
	}

	err = DeleteBucket(ctx, s3Client, bucketName)
	if err != nil {
		log.Printf("Error during bucket deletion: %v", err)
	}

	baseUrl, exists := os.LookupEnv("BASE_URL")
	if !exists {
		log.Fatal("Undefined base url")
	}

	err = InitBucket(ctx, s3Client, baseUrl, bucketName, awsRegion)
	if err != nil {
		log.Fatal("Failed to initialize bucket: ", err)
	}
}
