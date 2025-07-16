package api

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/HENNGE/snsclone-202506-golang-luca/dto"
	"github.com/HENNGE/snsclone-202506-golang-luca/service"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang-jwt/jwt/v5"

	"golang.org/x/oauth2"
)

type AuthHandler struct {
	BaseUrl  string
	Secret   string
	Config   oauth2.Config
	Provider *oidc.Provider
	Service  service.DefaultUserService
}

type AuthConfig struct {
	BaseUrl  string
	Secret   string
	Config   oauth2.Config
	Provider *oidc.Provider
}

func NewAuthHandler(service service.DefaultUserService, config *AuthConfig) *AuthHandler {
	return &AuthHandler{
		BaseUrl:  config.BaseUrl,
		Secret:   config.Secret,
		Config:   config.Config,
		Provider: config.Provider,
		Service:  service,
	}
}

func randString(nByte int) (string, error) {
	b := make([]byte, nByte)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func setCallbackCookie(w http.ResponseWriter, r *http.Request, name, value string) {
	c := &http.Cookie{
		Name:     name,
		Value:    value,
		MaxAge:   int(time.Hour.Seconds()),
		Secure:   r.TLS != nil,
		HttpOnly: false,
	}
	http.SetCookie(w, c)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	// Create anti-forgery tokens
	state, err := randString(16)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	// Save the token in a cookie
	setCallbackCookie(w, r, "state", state)

	// Send authentication request to Google
	// On success Google will redirect to the callback endpoint
	http.Redirect(w, r, h.Config.AuthCodeURL(state), http.StatusFound)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	c := &http.Cookie{
		Name:   cookieName,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(w, c)
	http.Redirect(w, r, h.BaseUrl, http.StatusFound)
}

func (h *AuthHandler) Callback(w http.ResponseWriter, r *http.Request) {
	state, err := r.Cookie("state")
	if err != nil {
		http.Error(w, "state not found", http.StatusBadRequest)
		return
	}
	if r.URL.Query().Get("state") != state.Value {
		http.Error(w, "state did not match", http.StatusBadRequest)
		return
	}

	oauth2Token, err := h.Config.Exchange(r.Context(), r.URL.Query().Get("code"))
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	userInfo, err := h.Provider.UserInfo(r.Context(), oauth2.StaticTokenSource(oauth2Token))
	if err != nil {
		http.Error(w, "Failed to get userinfo: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var UserClaims struct {
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}

	if err := userInfo.Claims(&UserClaims); err != nil {
		http.Error(w, "Failed to unmarshal userinfo claims: "+err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = h.Service.GetByID(r.Context(), userInfo.Subject)
	switch {
	case err == nil:
		// No error, continue with the callback workflow
	case !errors.Is(err, service.ErrUserNotFound):
		http.Error(w, "An internal server error occurred", http.StatusInternalServerError)
		return
	default:
		createUser := &dto.CreateUserRequest{
			ID:      userInfo.Subject,
			Name:    UserClaims.Name,
			Email:   userInfo.Email,
			Picture: UserClaims.Picture,
		}
		_, err := h.Service.Create(r.Context(), createUser)
		if err != nil {
			http.Error(w, "An internal server error occurred", http.StatusInternalServerError)
			return
		}
	}

	// Convert jwt secret into bytes
	jwtSecret := []byte(h.Secret)

	// Create claims to embed in jwt
	claims := AppClaims{
		UserID:   userInfo.Subject,
		UserName: UserClaims.Name,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "sns-clone",
		},
	}

	// Create new jwt
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(jwtSecret)
	if err != nil {
		http.Error(w, "Failed to sign token", http.StatusInternalServerError)
		return
	}

	// Embed jwt inside cookie
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    signedToken,
		Path:     "/",
		Expires:  time.Now().Add(time.Hour * 24),
		HttpOnly: true,
		Secure:   strings.Contains(h.BaseUrl, "https"),
		SameSite: http.SameSiteLaxMode,
	})

	http.Redirect(w, r, h.BaseUrl, http.StatusFound)
}
