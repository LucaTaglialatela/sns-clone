package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/HENNGE/snsclone-202506-golang-luca/entity"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type DefaultCommentRepository struct {
	DB        *dynamodb.Client
	TableName string
}

func NewDefaultCommentRepository(db *dynamodb.Client, tableName string) *DefaultCommentRepository {
	return &DefaultCommentRepository{
		DB:        db,
		TableName: tableName,
	}
}

type CommentRepository interface {
	Create(ctx context.Context, comment *entity.Comment) (*entity.Comment, error)
	GetByPostID(ctx context.Context, postId string) ([]*entity.Comment, error)
	GetRawByPostID(ctx context.Context, postId string) ([]map[string]types.AttributeValue, error)
	Get(ctx context.Context, postId, commentId string) (*entity.Comment, error)
	Update(ctx context.Context, postId, commentId, text string) (*entity.Comment, error)
	Delete(ctx context.Context, postId, commentId string) error
}

func (r *DefaultCommentRepository) Create(ctx context.Context, comment *entity.Comment) (*entity.Comment, error) {
	if comment == nil {
		return nil, fmt.Errorf("input comment cannot be nil")
	}

	av, err := attributevalue.MarshalMap(comment)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal user to DynamoDB attribute values: %w", err)
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(r.TableName),
	}

	_, err = r.DB.PutItem(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to put item (PK: %s) to DynamoDB: %w", comment.PK, err)
	}

	return comment, nil
}

func (r *DefaultCommentRepository) GetByPostID(ctx context.Context, postId string) ([]*entity.Comment, error) {
	var allRawItems []map[string]types.AttributeValue
	var comments []*entity.Comment

	partitionKey := "post#" + postId

	input := &dynamodb.QueryInput{
		TableName:              aws.String(r.TableName),
		KeyConditionExpression: aws.String("pk = :pk AND begins_with(sk, :sk_prefix)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk":        &types.AttributeValueMemberS{Value: partitionKey},
			":sk_prefix": &types.AttributeValueMemberS{Value: "comment#"},
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

	err := attributevalue.UnmarshalListOfMaps(allRawItems, &comments)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal DynamoDB items: %w", err)
	}

	return comments, nil
}

func (r *DefaultCommentRepository) GetRawByPostID(ctx context.Context, postId string) ([]map[string]types.AttributeValue, error) {
	var allRawItems []map[string]types.AttributeValue

	partitionKey := "post#" + postId

	input := &dynamodb.QueryInput{
		TableName:              aws.String(r.TableName),
		KeyConditionExpression: aws.String("pk = :pk AND begins_with(sk, :sk_prefix)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk":        &types.AttributeValueMemberS{Value: partitionKey},
			":sk_prefix": &types.AttributeValueMemberS{Value: "comment#"},
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

	return allRawItems, nil
}

func (r *DefaultCommentRepository) Get(ctx context.Context, postId, commentId string) (*entity.Comment, error) {
	key := map[string]types.AttributeValue{
		"pk": &types.AttributeValueMemberS{Value: fmt.Sprintf("post#%s", postId)},
		"sk": &types.AttributeValueMemberS{Value: fmt.Sprintf("comment#%s", commentId)},
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

	comment := entity.Comment{}
	err = attributevalue.UnmarshalMap(result.Item, &comment)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling item: %w", err)
	}

	return &comment, nil
}

func (r *DefaultCommentRepository) Update(ctx context.Context, postId, commentId, text string) (*entity.Comment, error) {
	key := map[string]types.AttributeValue{
		"pk": &types.AttributeValueMemberS{Value: fmt.Sprintf("post#%s", postId)},
		"sk": &types.AttributeValueMemberS{Value: fmt.Sprintf("comment#%s", commentId)},
	}

	updateExpression := ("SET #text = :text, #edited = :edited")
	expressionAttributeNames := map[string]string{
		"#text":   "text",
		"#edited": "edited",
	}
	expressionAttributeValues := map[string]types.AttributeValue{
		":text":   &types.AttributeValueMemberS{Value: text},
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

	updatedComment := entity.Comment{}
	err = attributevalue.UnmarshalMap(result.Attributes, &updatedComment)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling item: %w", err)
	}

	return &updatedComment, nil
}

func (r *DefaultCommentRepository) Delete(ctx context.Context, postId, commentId string) error {
	key := map[string]types.AttributeValue{
		"pk": &types.AttributeValueMemberS{Value: fmt.Sprintf("post#%s", postId)},
		"sk": &types.AttributeValueMemberS{Value: fmt.Sprintf("comment#%s", commentId)},
	}

	dynamoDbInput := &dynamodb.DeleteItemInput{
		TableName:    aws.String(r.TableName),
		Key:          key,
		ReturnValues: types.ReturnValueAllOld,
	}

	result, err := r.DB.DeleteItem(ctx, dynamoDbInput)
	if err != nil {
		return fmt.Errorf("failed to delete item: %w", err)
	}

	deletedComment := entity.Comment{}
	err = attributevalue.UnmarshalMap(result.Attributes, &deletedComment)
	if err != nil {
		return fmt.Errorf("error unmarshalling item: %w", err)
	}

	return nil
}
