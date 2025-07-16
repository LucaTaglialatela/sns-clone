package repository

import (
	"context"
	"fmt"

	"github.com/HENNGE/snsclone-202506-golang-luca/entity"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go/aws"
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) (*entity.User, error)
	GetByID(ctx context.Context, id string) (*entity.User, error)
	GetAll(ctx context.Context) ([]*entity.User, error)
	GetFollowing(ctx context.Context) ([]*entity.Follow, error)
	Follow(ctx context.Context, follow *entity.Follow) (*entity.Follow, error)
	Unfollow(ctx context.Context, unfollow *entity.Unfollow) (error)
}

type DefaultUserRepository struct {
	DB        *dynamodb.Client
	TableName string
}

func NewDefaultUserRepository(db *dynamodb.Client, tableName string) *DefaultUserRepository {
	return &DefaultUserRepository{
		DB:        db,
		TableName: tableName,
	}
}

func (r *DefaultUserRepository) Create(ctx context.Context, user *entity.User) (*entity.User, error) {
	if user == nil {
		return nil, fmt.Errorf("input user cannot be nil")
	}

	av, err := attributevalue.MarshalMap(user)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal user to DynamoDB attribute values: %w", err)
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(r.TableName),
	}

	_, err = r.DB.PutItem(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to put item (PK: %s) to DynamoDB: %w", user.PK, err)
	}

	return user, nil
}

func (r *DefaultUserRepository) GetByID(ctx context.Context, id string) (*entity.User, error) {
	key := map[string]types.AttributeValue{
		"pk": &types.AttributeValueMemberS{Value: "user"},
		"sk": &types.AttributeValueMemberS{Value: id},
	}

	input := &dynamodb.GetItemInput{
		TableName: aws.String(r.TableName),
		Key:       key,
	}

	result, err := r.DB.GetItem(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get item from DynamoDB (pk: user, sk: %s): %w", id, err)
	}

	if result.Item == nil {
		return nil, nil
	}

	var item entity.User
	err = attributevalue.UnmarshalMap(result.Item, &item)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal DynamoDB item (pk: user, sk: %s): %w", id, err)
	}

	return &item, nil
}

func (r *DefaultUserRepository) GetAll(ctx context.Context) ([]*entity.User, error) {
	var allRawItems []map[string]types.AttributeValue
	var users []*entity.User

	input := &dynamodb.QueryInput{
		TableName:              aws.String(r.TableName),
		KeyConditionExpression: aws.String("pk = :pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: "user"},
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

	err := attributevalue.UnmarshalListOfMaps(allRawItems, &users)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal DynamoDB items: %w", err)
	}

	return users, nil
}

func (r *DefaultUserRepository) GetFollowing(ctx context.Context, userID string) ([]*entity.Follow, error) {
    var allRawItems []map[string]types.AttributeValue
    var followers []*entity.Follow

    partitionKey := "user#" + userID

    input := &dynamodb.QueryInput{
        TableName:              aws.String(r.TableName),
        KeyConditionExpression: aws.String("pk = :pk AND begins_with(sk, :sk_prefix)"),
        ExpressionAttributeValues: map[string]types.AttributeValue{
            ":pk":        &types.AttributeValueMemberS{Value: partitionKey},
            ":sk_prefix": &types.AttributeValueMemberS{Value: "follower#"},
        },
        ScanIndexForward: aws.Bool(false),
    }

    paginator := dynamodb.NewQueryPaginator(r.DB, input)

    for paginator.HasMorePages() {
        page, err := paginator.NextPage(ctx)
        if err != nil {
            return nil, fmt.Errorf("failed to get next page of query results for user %s: %w", userID, err)
        }
        allRawItems = append(allRawItems, page.Items...)
    }

    err := attributevalue.UnmarshalListOfMaps(allRawItems, &followers)
    if err != nil {
        return nil, fmt.Errorf("failed to unmarshal DynamoDB items into Follow struct: %w", err)
    }

    return followers, nil
}

func (r *DefaultUserRepository) Follow(ctx context.Context, follow *entity.Follow) (*entity.Follow, error) {
	if follow == nil {
		return nil, fmt.Errorf("input follow cannot be nil")
	}

	av, err := attributevalue.MarshalMap(follow)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal follow to DynamoDB attribute values: %w", err)
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(r.TableName),
	}

	_, err = r.DB.PutItem(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to put item (PK: %s) to DynamoDB: %w", follow.PK, err)
	}

	return follow, nil
}

func (r *DefaultUserRepository) Unfollow(ctx context.Context, unfollow *entity.Unfollow) error {
    if unfollow == nil {
        return fmt.Errorf("input unfollow cannot be nil")
    }

    key, err := attributevalue.MarshalMap(map[string]interface{}{
        "pk": unfollow.PK,
        "sk": unfollow.SK,
    })
    if err != nil {
        return fmt.Errorf("failed to marshal key: %w", err)
    }

    input := &dynamodb.DeleteItemInput{
        Key:       key,
        TableName: aws.String(r.TableName),
    }

    _, err = r.DB.DeleteItem(ctx, input)
    if err != nil {
        return fmt.Errorf("failed to delete item (PK: %s) from DynamoDB: %w", unfollow.PK, err)
    }

    return nil
}
