package api

import (
	"net/http"
)

func NewRouter(h *Handlers, jwtSecret string) *http.ServeMux {
	mux := http.NewServeMux()

	authMiddleware := AuthMiddleware(jwtSecret)

	mux.HandleFunc("/", h.ServeHandler.Serve)

	mux.HandleFunc("GET /ping", h.PingHandler.Ping)

	mux.Handle("GET /me", authMiddleware(http.HandlerFunc(h.UserHandler.Me)))
	mux.HandleFunc("POST /users", h.UserHandler.Create)
	mux.Handle("GET /users/{id}", authMiddleware(http.HandlerFunc(h.UserHandler.GetByID)))
	mux.HandleFunc("GET /users", h.UserHandler.GetAll)
	mux.Handle("POST /users/follow", authMiddleware(http.HandlerFunc(h.UserHandler.Follow)))
	mux.Handle("DELETE /users/unfollow", authMiddleware(http.HandlerFunc(h.UserHandler.Unfollow)))

	mux.Handle("POST /posts", authMiddleware(http.HandlerFunc(h.PostHandler.Create)))
	mux.Handle("GET /posts", authMiddleware(http.HandlerFunc(h.PostHandler.GetAll)))
	mux.Handle("GET /posts/{user_id}", authMiddleware(http.HandlerFunc(h.PostHandler.GetByUserID)))
	mux.Handle("PUT /users/{user_id}/posts/{post_id}", authMiddleware(http.HandlerFunc(h.PostHandler.Update)))
	mux.Handle("DELETE /users/{user_id}/posts/{post_id}", authMiddleware(http.HandlerFunc(h.PostHandler.Delete)))

	mux.Handle("POST /posts/{post_id}/comments", authMiddleware(http.HandlerFunc(h.CommentHandler.Create)))
	mux.Handle("GET /posts/{post_id}/comments", authMiddleware(http.HandlerFunc(h.CommentHandler.GetByPostID)))
	mux.Handle("PUT /users/{user_id}/posts/{post_id}/comments/{comment_id}", authMiddleware(http.HandlerFunc(h.CommentHandler.Update)))
	mux.Handle("DELETE /users/{user_id}/posts/{post_id}/comments/{comment_id}", authMiddleware(http.HandlerFunc(h.CommentHandler.Delete)))

	mux.Handle("POST /presign", authMiddleware(http.HandlerFunc(h.S3PresignHandler.Upload)))

	mux.HandleFunc("/auth/google/login", h.AuthHandler.Login)
	mux.HandleFunc("/auth/google/callback", h.AuthHandler.Callback)
	mux.HandleFunc("/auth/logout", h.AuthHandler.Logout)

	mux.Handle("/events", authMiddleware(http.HandlerFunc(h.Broker.ServeHTTP)))

	return mux
}
