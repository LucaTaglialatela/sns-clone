package service

import (
	"context"
	"errors"

	"github.com/HENNGE/snsclone-202506-golang-luca/dto"
	"github.com/HENNGE/snsclone-202506-golang-luca/entity"
	"github.com/HENNGE/snsclone-202506-golang-luca/repository"
)

var ErrUserNotFound = errors.New("user not found")

type UserService interface {
	Create(ctx context.Context, request *dto.CreateUserRequest) (*dto.User, error)
	GetByID(ctx context.Context, id string) (*dto.User, error)
	GetAll(ctx context.Context) ([]*dto.User, error)
	GetFollowing(ctx context.Context, userID string) ([]*dto.Follow, error)
	Follow(ctx context.Context, userID string, request *dto.FollowRequest) (*dto.Follow, error)
	Unfollow(ctx context.Context, userID string, request *dto.UnfollowRequest) (error)
}

type DefaultUserService struct {
	repository repository.DefaultUserRepository
}

func NewDefaultUserService(repository repository.DefaultUserRepository) *DefaultUserService {
	return &DefaultUserService{
		repository: repository,
	}
}

func (s *DefaultUserService) Create(ctx context.Context, request *dto.CreateUserRequest) (*dto.User, error) {
	user, err := entity.NewUser(request.ID, request.Name, request.Email, request.Picture)
	if err != nil {
		return nil, err
	}

	createdUser, err := s.repository.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	userDto := new(dto.User)
	userDto.FromEntity(createdUser)

	return userDto, nil
}

func (s *DefaultUserService) GetByID(ctx context.Context, id string) (*dto.User, error) {
	user, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, ErrUserNotFound
	}

	userDto := new(dto.User)
	userDto.FromEntity(user)

	return userDto, nil
}

func (s *DefaultUserService) GetAll(ctx context.Context) ([]*dto.User, error) {
	users, err := s.repository.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	userDtos := make([]*dto.User, 0, len(users))
	for _, user := range users {
		userDto := new(dto.User)
		userDto.FromEntity(user)
		userDtos = append(userDtos, userDto)
	}

	return userDtos, nil
}

func (s *DefaultUserService) GetFollowing(ctx context.Context, userID string) (*dto.Following, error) {
	following, err := s.repository.GetFollowing(ctx, userID)
	if err != nil {
		return nil, err
	}

	followingDto := new(dto.Following)
	followingDto.FromEntity(following)

	return followingDto, nil
}

func (s *DefaultUserService) Follow(ctx context.Context, userID string, request *dto.FollowRequest) (*dto.Follow, error) {
	follow, err := entity.NewFollow(userID, request.FollowingID)
	if err != nil {
		return nil, err
	}

	createdFollow, err := s.repository.Follow(ctx, follow)
	if err != nil {
		return nil, err
	}

	followDto := new(dto.Follow)
	followDto.FromEntity(createdFollow)

	return followDto, nil
}

func (s *DefaultUserService) Unfollow(ctx context.Context, userID string, request *dto.UnfollowRequest) (error) {
	unfollow, err := entity.NewUnfollow(userID, request.UnfollowingID)
	if err != nil {
		return err
	}

	err = s.repository.Unfollow(ctx, unfollow)
	if err != nil {
		return err
	}

	return nil
}
