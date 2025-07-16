package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

type AppClaims struct {
	UserID string `json:"user_id"`
	UserName string `json:"user_name"`
	jwt.RegisteredClaims
}

type contextKey string

const userClaimsKey contextKey = "userClaims"
const cookieName string = "auth_token"

// Authentication middleware to check for a valid JWT in the "auth_token" cookie
func AuthMiddleware(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get the cookie from the request
			cookie, err := r.Cookie(cookieName)
			if err != nil {
				// If the cookie is not found, return an unauthorized error
				http.Error(w, "Unauthorized: No token provided", http.StatusUnauthorized)
				return
			}

			// Parse and validate the JWT from the cookie value
			tokenString := cookie.Value
			claims := &AppClaims{}

			token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(jwtSecret), nil
			})

			if err != nil || !token.Valid {
				// If the token is invalid (e.g., expired, bad signature), return an error
				http.Error(w, "Unauthorized: Invalid token", http.StatusUnauthorized)
				return
			}

			// If the token is valid, put the claims into the request context
			ctx := context.WithValue(r.Context(), userClaimsKey, claims)

			// Call the next handler in the chain with the new context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
