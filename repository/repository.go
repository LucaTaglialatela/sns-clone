package repository

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Repositories struct {
	UserRepository    *DefaultUserRepository
	PostRepository    *DefaultPostRepository
	CommentRepository *DefaultCommentRepository
}

func InitRepositories(db *dynamodb.Client, s3Client *s3.Client, s3PresignClient *s3.PresignClient, tableName, bucketName string) *Repositories {
	commentRepository := NewDefaultCommentRepository(db, tableName)
	return &Repositories{
		UserRepository:    NewDefaultUserRepository(db, tableName),
		PostRepository:    NewDefaultPostRepository(db, s3Client, s3PresignClient, commentRepository, tableName, bucketName),
		CommentRepository: commentRepository,
	}
}
