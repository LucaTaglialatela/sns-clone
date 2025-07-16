package service

import (
	"context"
	"fmt"

	"github.com/HENNGE/snsclone-202506-golang-luca/dto"
	"github.com/HENNGE/snsclone-202506-golang-luca/entity"
	"github.com/HENNGE/snsclone-202506-golang-luca/repository"
)

type PostService interface {
	Create(ctx context.Context, userID, userName string, request *dto.CreatePostRequest) (*dto.Post, error)
	GetAll(ctx context.Context) ([]*dto.Post, error)
	GetByUserID(ctx context.Context, userId string) ([]*dto.Post, error)
	Update(ctx context.Context, userId, postId, text, image string) (*dto.Post, error)
	Delete(ctx context.Context, userId, postId string) (error)
}

type DefaultPostService struct {
	repository repository.DefaultPostRepository
}

func NewDefaultPostService(repository repository.DefaultPostRepository) *DefaultPostService {
	return &DefaultPostService{
		repository: repository,
	}
}

func (s *DefaultPostService) Create(ctx context.Context, userID, userName string, request *dto.CreatePostRequest) (*dto.Post, error) {
	post, err := entity.NewPost(userID, userName, request.Text, request.Image)
	if err != nil {
		return nil, fmt.Errorf("failed to create post entity: %w", err)
	}

	createdPost, err := s.repository.Create(ctx, post)
	if err != nil {
		return nil, err
	}

	postDto := new(dto.Post)
	postDto.FromEntity(createdPost)

	return postDto, nil
}

func (s *DefaultPostService) GetAll(ctx context.Context) ([]*dto.Post, error) {
	posts, err := s.repository.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	postDtos := make([]*dto.Post, 0, len(posts))
	for _, post := range posts {
		postDto := new(dto.Post)
		postDto.FromEntity(post)
		postDtos = append(postDtos, postDto)
	}

	return postDtos, nil
}

func (s *DefaultPostService) GetByUserID(ctx context.Context, userId string) ([]*dto.Post, error) {
	posts, err := s.repository.GetByUserID(ctx, userId)
	if err != nil {
		return nil, err
	}

	postDtos := make([]*dto.Post, 0, len(posts))
	for _, post := range posts {
		postDto := new(dto.Post)
		postDto.FromEntity(post)
		postDtos = append(postDtos, postDto)
	}

	return postDtos, nil
}

func (s *DefaultPostService) Update(ctx context.Context, userId, postId string, request *dto.UpdatePostRequest) (*dto.Post, error) {
	post, err := s.repository.Get(ctx, userId, postId)
	if err != nil {
		return nil, fmt.Errorf("cannot find post to update: %w", err)
	}

	updatedPost, err := s.repository.Update(ctx, userId, postId, request.Text, request.Image)
	if err != nil {
		return nil, fmt.Errorf("failed to update post: %w", err)
	}

	// If the image was updated, delete the old image
	if post.Image != request.Image && post.Image != "" {
		err = s.repository.DeleteImage(ctx, post.Image)
		if err != nil {
			return nil, fmt.Errorf("failed to remove old image: %w", err)
		}
	}

	postDto := new(dto.Post)
	postDto.FromEntity(updatedPost)

	return postDto, nil
}

func (s *DefaultPostService) Delete(ctx context.Context, userId, postId string) (error) {
	_, err := s.repository.Get(ctx, userId, postId)
	if err != nil {
		return fmt.Errorf("cannot find post to delete: %w", err)
	}

	err = s.repository.Delete(ctx, userId, postId)
	if err != nil {
		return fmt.Errorf("failed to delete post: %w", err)
	}

	return nil
}
