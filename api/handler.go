package api

import (
	"context"
	"net/http"

	"github.com/HENNGE/snsclone-202506-golang-luca/entity"
	"github.com/HENNGE/snsclone-202506-golang-luca/service"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Handlers struct {
	PingHandler      *PingHandler
	UserHandler      *UserHandler
	PostHandler      *PostHandler
	CommentHandler   *CommentHandler
	AuthHandler      *AuthHandler
	ServeHandler     *ServeHandler
	S3PresignHandler *S3PresignHandler
	Broker           *entity.Broker
}

func InitHandlers(services *service.Services, authConfig *AuthConfig, fs http.Handler, client *s3.PresignClient) *Handlers {
	broker := entity.NewBroker()
	return &Handlers{
		PingHandler:      NewPingHandler(),
		UserHandler:      NewUserHandler(*services.UserService),
		PostHandler:      NewPostHandler(*services.PostService, broker),
		CommentHandler:   NewCommentHandler(*services.CommentService),
		AuthHandler:      NewAuthHandler(*services.UserService, authConfig),
		ServeHandler:     NewServeHandler(fs),
		S3PresignHandler: NewS3PresignHandler(client),
		Broker:           broker,
	}
}

func IsAuthorized(ctx context.Context, userId string) bool {
	claims, ok := ctx.Value(userClaimsKey).(*AppClaims)
	if !ok {
		return false
	}

	if claims.UserID != userId {
		return false
	}

	return true
}
