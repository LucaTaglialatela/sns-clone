package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/HENNGE/snsclone-202506-golang-luca/entity"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type PostRepository interface {
	Create(ctx context.Context, post *entity.Post) (*entity.Post, error)
	GetAll(ctx context.Context) ([]*entity.Post, error)
	GetByUserID(ctx context.Context, userID string) ([]*entity.Post, error)
	Get(ctx context.Context, userId, postId string) (*entity.Post, error)
	Update(ctx context.Context, userId, postId, text, image string) (*entity.Post, error)
	Delete(ctx context.Context, userId, postId string) error
	DeleteImage(ctx context.Context, imageKey string) error
}

type DefaultPostRepository struct {
	DB                *dynamodb.Client
	S3                *s3.Client
	S3PS              *s3.PresignClient
	CommentRepository *DefaultCommentRepository
	TableName         string
	BucketName        string
}

type BatchWriteItemConfig struct {
	BatchSize      int
	MaxRetries     int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
}

func NewDefaultPostRepository(db *dynamodb.Client, s3Client *s3.Client, s3PresignClient *s3.PresignClient, commentRepository *DefaultCommentRepository, tableName, bucketName string) *DefaultPostRepository {
	return &DefaultPostRepository{
		DB:                db,
		S3:                s3Client,
		S3PS:              s3PresignClient,
		CommentRepository: commentRepository,
		TableName:         tableName,
		BucketName:        bucketName,
	}
}

func (r *DefaultPostRepository) Create(ctx context.Context, post *entity.Post) (*entity.Post, error) {
	if post == nil {
		return nil, fmt.Errorf("input post cannot be nil")
	}

	err := r.ValidateImage(ctx, post.Image)
	if err != nil {
		return nil, fmt.Errorf("failed to validate image: %w", err)
	}

	av, err := attributevalue.MarshalMap(post)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal user to DynamoDB attribute values: %w", err)
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(r.TableName),
	}

	_, err = r.DB.PutItem(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to put item (PK: %s) to DynamoDB: %w", post.PK, err)
	}

	err = r.SetImageURL(ctx, post)
	if err != nil {
		return nil, fmt.Errorf("failed to set image url: %w", err)
	}

	return post, nil
}

func (r *DefaultPostRepository) GetAll(ctx context.Context) ([]*entity.Post, error) {
	var allRawItems []map[string]types.AttributeValue
	var posts []*entity.Post

	input := &dynamodb.QueryInput{
		TableName:              aws.String(r.TableName),
		IndexName:              aws.String("gsi1"),
		KeyConditionExpression: aws.String("gsi1_pk = :pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: "timeline"},
		},
		ScanIndexForward: aws.Bool(false),
	}

	paginator := dynamodb.NewQueryPaginator(r.DB, input)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get next page of query results: %w", err)
		}
		allRawItems = append(allRawItems, page.Items...)
	}

	err := attributevalue.UnmarshalListOfMaps(allRawItems, &posts)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal DynamoDB items: %w", err)
	}

	for _, post := range posts {
		err = r.SetImageURL(ctx, post)
		if err != nil {
			return nil, fmt.Errorf("failed to set image url: %w", err)
		}
	}

	return posts, nil
}

func (r *DefaultPostRepository) GetByUserID(ctx context.Context, userID string) ([]*entity.Post, error) {
	var allRawItems []map[string]types.AttributeValue
	var posts []*entity.Post

	partitionKey := "user#" + userID

	input := &dynamodb.QueryInput{
		TableName:              aws.String(r.TableName),
		KeyConditionExpression: aws.String("pk = :pk AND begins_with(sk, :sk_prefix)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk":        &types.AttributeValueMemberS{Value: partitionKey},
			":sk_prefix": &types.AttributeValueMemberS{Value: "post#"},
		},
		ScanIndexForward: aws.Bool(false),
	}

	paginator := dynamodb.NewQueryPaginator(r.DB, input)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get next page of query results: %w", err)
		}
		allRawItems = append(allRawItems, page.Items...)
	}

	err := attributevalue.UnmarshalListOfMaps(allRawItems, &posts)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal DynamoDB items: %w", err)
	}

	for _, post := range posts {
		err = r.SetImageURL(ctx, post)
		if err != nil {
			return nil, fmt.Errorf("failed to set image url: %w", err)
		}
	}

	return posts, nil
}

func (r *DefaultPostRepository) Get(ctx context.Context, userId, postId string) (*entity.Post, error) {
	key := map[string]types.AttributeValue{
		"pk": &types.AttributeValueMemberS{Value: fmt.Sprintf("user#%s", userId)},
		"sk": &types.AttributeValueMemberS{Value: fmt.Sprintf("post#%s", postId)},
	}

	input := &dynamodb.GetItemInput{
		TableName: &r.TableName,
		Key:       key,
	}

	result, err := r.DB.GetItem(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("error getting item: %w", err)
	}

	if result.Item == nil {
		return nil, fmt.Errorf("item not found")
	}

	post := entity.Post{}
	err = attributevalue.UnmarshalMap(result.Item, &post)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling item: %w", err)
	}

	return &post, nil
}

func (r *DefaultPostRepository) Update(ctx context.Context, userId, postId, text, image string) (*entity.Post, error) {
	key := map[string]types.AttributeValue{
		"pk": &types.AttributeValueMemberS{Value: fmt.Sprintf("user#%s", userId)},
		"sk": &types.AttributeValueMemberS{Value: fmt.Sprintf("post#%s", postId)},
	}

	updateExpression := ("SET #text = :text, #image = :image, #edited = :edited")
	expressionAttributeNames := map[string]string{
		"#text":   "text",
		"#image":  "image",
		"#edited": "edited",
	}
	expressionAttributeValues := map[string]types.AttributeValue{
		":text":   &types.AttributeValueMemberS{Value: text},
		":image":  &types.AttributeValueMemberS{Value: image},
		":edited": &types.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339Nano)},
	}

	input := &dynamodb.UpdateItemInput{
		TableName:                 &r.TableName,
		Key:                       key,
		UpdateExpression:          &updateExpression,
		ExpressionAttributeNames:  expressionAttributeNames,
		ExpressionAttributeValues: expressionAttributeValues,
		ReturnValues:              types.ReturnValueAllNew,
	}

	result, err := r.DB.UpdateItem(ctx, input)
	if err != nil {
		fmt.Println("Error updating item: ", err)
		return nil, fmt.Errorf("error updating item: %w", err)
	}

	updatedPost := entity.Post{}
	err = attributevalue.UnmarshalMap(result.Attributes, &updatedPost)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling item: %w", err)
	}

	err = r.SetImageURL(ctx, &updatedPost)
	if err != nil {
		return nil, fmt.Errorf("failed to set image url: %w", err)
	}

	return &updatedPost, nil
}

func (r *DefaultPostRepository) Delete(ctx context.Context, userId, postId string) error {
	// Check if post exists in the database
	post, err := r.Get(ctx, userId, postId)
	if err != nil {
		return fmt.Errorf("failed to get post to delete: %w", err)
	}

	// Get comments related to post from the database
	comments, err := r.CommentRepository.GetRawByPostID(ctx, postId)
	if err != nil {
		return fmt.Errorf("failed to get post comments: %w", err)
	}

	config := &BatchWriteItemConfig{
		BatchSize:      25,
		MaxRetries:     5,
		InitialBackoff: 100 * time.Millisecond,
		MaxBackoff:     5 * time.Second,
	}

	// Attempt to delete the comments first
	err = r.DeleteComments(ctx, postId, comments, config)
	if err != nil {
		return fmt.Errorf("failed to batch delete comments: %w", err)
	}

	key := map[string]types.AttributeValue{
		"pk": &types.AttributeValueMemberS{Value: fmt.Sprintf("user#%s", userId)},
		"sk": &types.AttributeValueMemberS{Value: fmt.Sprintf("post#%s", postId)},
	}

	dynamoDbInput := &dynamodb.DeleteItemInput{
		TableName:    aws.String(r.TableName),
		Key:          key,
		ReturnValues: types.ReturnValueAllOld,
	}

	// If the post contained an image, delete the image from S3
	err = r.DeleteImage(ctx, post.Image)
	if err != nil {
		return fmt.Errorf("error deleting s3 item: %w", err)
	}

	// If all comments were deleted, proceed with deleting the post
	_, err = r.DB.DeleteItem(ctx, dynamoDbInput)
	if err != nil {
		return fmt.Errorf("failed to delete item: %w", err)
	}

	return nil
}

func (r *DefaultPostRepository) DeleteComments(ctx context.Context, postId string, comments []map[string]types.AttributeValue, config *BatchWriteItemConfig) error {
	remainingComments := make([]map[string]types.AttributeValue, len(comments))
	copy(remainingComments, comments)

	var unprocessedComments []types.WriteRequest

	for len(remainingComments) > 0 {
		currentBatch := []types.WriteRequest{}
		// Create a batch of delete write requests of size config.BatchSize
		for i := 0; i < len(remainingComments) && i < config.BatchSize; i++ {
			comment := remainingComments[i]
			writeRequests := types.WriteRequest{
				DeleteRequest: &types.DeleteRequest{
					Key: map[string]types.AttributeValue{
						"pk": comment["pk"],
						"sk": comment["sk"],
					},
				},
			}
			currentBatch = append(currentBatch, writeRequests)
		}

		input := &dynamodb.BatchWriteItemInput{
			RequestItems: map[string][]types.WriteRequest{
				r.TableName: currentBatch,
			},
		}

		// Execute the batch write requests
		output, err := r.DB.BatchWriteItem(ctx, input)
		if err != nil {
			return fmt.Errorf("DynamoDB BatchWriteItem failed: %w", err)
		}

		remainingComments = remainingComments[len(currentBatch):]

		// Collect any items that DynamoDB could not process.
		if unprocessed := output.UnprocessedItems[r.TableName]; len(unprocessed) > 0 {
			unprocessedComments = append(unprocessed, unprocessed...)
		}
	}

	retries := 0
	for len(unprocessedComments) > 0 && retries < config.MaxRetries {
		// Create a new batch from the unprocessedComments
		var currentBatch []types.WriteRequest
		if len(unprocessedComments) > config.BatchSize {
			currentBatch = unprocessedComments[:config.BatchSize]
		} else {
			currentBatch = unprocessedComments
		}
		
		// Keep track of the remaining unprocessedComments
		remainingUnprocessed := unprocessedComments[len(currentBatch):]

		input := &dynamodb.BatchWriteItemInput{
			RequestItems: map[string][]types.WriteRequest{
				r.TableName: currentBatch,
			},
		}

		output, err := r.DB.BatchWriteItem(ctx, input)
		if err != nil {
			return fmt.Errorf("DynamoDB BatchWriteItem failed: %w", err)
		}

		// Check if the current batch has unprocessed items again
		unprocessedAgain := output.UnprocessedItems[r.TableName]

		// If yes, put them in the front of the unprocessed items so we can retry
		if len(unprocessedAgain) > 0 {
			unprocessedComments = append(unprocessedAgain, remainingUnprocessed...)
			retries++
			
			// Apply exponential backoff
			sleepTime := min(config.InitialBackoff * time.Duration(math.Pow(2, float64(retries-1))), config.MaxBackoff)
			time.Sleep(sleepTime)
		} else {
			// All unprocessed items were successfully processed this time
			unprocessedComments = remainingUnprocessed
			retries = 0
		}
	}

	if len(unprocessedComments) > 0 {
		return fmt.Errorf("failed to delete all comments after %d retries, %d items remain unprocessed", config.MaxRetries, len(unprocessedComments))
	}

	return nil
}

func (r *DefaultPostRepository) ValidateImage(ctx context.Context, imageKey string) error {
	if imageKey == "" {
		return nil
	}
	
	input := &s3.HeadObjectInput{
		Bucket: aws.String(r.BucketName),
		Key:    aws.String(imageKey),
	}

	_, err := r.S3.HeadObject(ctx, input)
	if err != nil {
		return err
	}

	return nil
}

func (r *DefaultPostRepository) SetImageURL(ctx context.Context, post *entity.Post) error {
	if post == nil || post.Image == "" {
		return nil
	}

	presignRequest, err := r.S3PS.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(r.BucketName),
		Key:    aws.String(post.Image),
	}, s3.WithPresignExpires(*aws.Duration(15 * time.Minute)))

	if err != nil {
		log.Printf("Couldn't get presigned URL for key %s. Here's why: %v\n", post.Image, err)
		post.ImageURL = nil
		return err
	}

	post.ImageURL = &presignRequest.URL
	return nil
}

func (r *DefaultPostRepository) DeleteImage(ctx context.Context, imageKey string) error {
	if imageKey == "" {
		return nil
	}

	err := r.ValidateImage(ctx, imageKey)
	if err != nil {
		var notFoundErr *s3Types.NotFound
		if errors.As(err, &notFoundErr) {
			return nil
		} else {
			return err
		}
	}

	s3Input := &s3.DeleteObjectInput{
		Bucket: aws.String(r.BucketName),
		Key:    aws.String(imageKey),
	}

	_, err = r.S3.DeleteObject(ctx, s3Input)
	if err != nil {
		return err
	}

	return nil
}
