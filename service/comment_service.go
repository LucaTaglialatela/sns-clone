package service

import (
	"context"
	"fmt"

	"github.com/HENNGE/snsclone-202506-golang-luca/dto"
	"github.com/HENNGE/snsclone-202506-golang-luca/entity"
	"github.com/HENNGE/snsclone-202506-golang-luca/repository"
)

type CommentService interface {
	Create(ctx context.Context, postId, userId, userName string, request *dto.SaveCommentRequest) (*dto.Comment, error)
	GetByPostID(ctx context.Context, postId string) ([]*dto.Comment, error)
	Update(ctx context.Context, postId, commentId string, request *dto.SaveCommentRequest) (*entity.Comment, error)
	Delete(ctx context.Context, postId, commentId string) (error)
}

type DefaultCommentService struct {
	repository repository.DefaultCommentRepository
}

func NewDefaultCommentService(repository repository.DefaultCommentRepository) *DefaultCommentService {
	return &DefaultCommentService{
		repository: repository,
	}
}

func (s *DefaultCommentService) Create(ctx context.Context, postId, userId, userName string, request *dto.SaveCommentRequest) (*dto.Comment, error) {
	comment, err := entity.NewComment(postId, userId, userName, request.Text)
	if err != nil {
		return nil, fmt.Errorf("failed to create comment entity: %w", err)
	}

	createdComment, err := s.repository.Create(ctx, comment)
	if err != nil {
		return nil, err
	}

	commentDto := new(dto.Comment)
	commentDto.FromEntity(createdComment)

	return commentDto, nil
}

func (s *DefaultCommentService) GetByPostID(ctx context.Context, postId string) ([]*dto.Comment, error) {
	comments, err := s.repository.GetByPostID(ctx, postId)
	if err != nil {
		return nil, err
	}

	commentDtos := make([]*dto.Comment, 0, len(comments))
	for _, comment := range comments {
		commentDto := new(dto.Comment)
		commentDto.FromEntity(comment)
		commentDtos = append(commentDtos, commentDto)
	}

	return commentDtos, nil
}

func (s *DefaultCommentService) Update(ctx context.Context, postId, commentId string, request *dto.SaveCommentRequest) (*entity.Comment, error) {
	_, err := s.repository.Get(ctx, postId, commentId)
	if err != nil {
		return nil, fmt.Errorf("cannot find comment to update: %w", err)
	}

	updatedComment, err := s.repository.Update(ctx, postId, commentId, request.Text)
	if err != nil {
		return nil, fmt.Errorf("failed to update comment: %w", err)
	}

	return updatedComment, nil
}

func (s *DefaultCommentService) Delete(ctx context.Context, postId, commentId string) (error) {
	_, err := s.repository.Get(ctx, postId, commentId)
	if err != nil {
		return fmt.Errorf("cannot find comment to delete: %w", err)
	}

	err = s.repository.Delete(ctx, postId, commentId)
	if err != nil {
		return fmt.Errorf("failed to delete comment: %w", err)
	}

	return nil
}
