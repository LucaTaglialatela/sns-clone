package service

import "github.com/HENNGE/snsclone-202506-golang-luca/repository"

type Services struct {
	UserService    *DefaultUserService
	PostService    *DefaultPostService
	CommentService *DefaultCommentService
}

func InitServices(repositories *repository.Repositories) *Services {
	return &Services{
		UserService:    NewDefaultUserService(*repositories.UserRepository),
		PostService:    NewDefaultPostService(*repositories.PostRepository),
		CommentService: NewDefaultCommentService(*repositories.CommentRepository),
	}
}
