package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/HENNGE/snsclone-202506-golang-luca/dto"
	"github.com/HENNGE/snsclone-202506-golang-luca/service"
)

type UserHandler struct {
	service service.DefaultUserService
}

func NewUserHandler(service service.DefaultUserService) *UserHandler {
	return &UserHandler{
		service: service,
	}
}

func (h *UserHandler) Me(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(userClaimsKey).(*AppClaims)
	if !ok {
		http.Error(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	user, err := h.service.GetByID(r.Context(), claims.UserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} 

	following, err := h.service.GetFollowing(r.Context(), claims.UserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := dto.UserProfileResponse{
		ID:        claims.UserID,
		Name:      user.Name,
		Following: following.FollowingIDs,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	request := dto.CreateUserRequest{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if request.Name == "" {
		log.Printf("name is empty")
		http.Error(w, "name cannot be empty", http.StatusBadRequest)
		return
	}

	user, err := h.service.Create(r.Context(), &request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if id == "" {
		log.Printf("id is empty")
		http.Error(w, "id cannot be empty", http.StatusBadRequest)
		return
	}

	user, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	users, err := h.service.GetAll(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
}

func (h *UserHandler) Follow(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(userClaimsKey).(*AppClaims)
	if !ok {
		http.Error(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	request := dto.FollowRequest{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if request.FollowingID == "" {
		log.Printf("following id is empty")
		http.Error(w, "following id cannot be empty", http.StatusBadRequest)
		return
	}

	follow, err := h.service.Follow(r.Context(), claims.UserID, &request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(follow)
}

func (h *UserHandler) Unfollow(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(userClaimsKey).(*AppClaims)
	if !ok {
		http.Error(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	request := dto.UnfollowRequest{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if request.UnfollowingID == "" {
		log.Printf("following id is empty")
		http.Error(w, "following id cannot be empty", http.StatusBadRequest)
		return
	}

	err := h.service.Unfollow(r.Context(), claims.UserID, &request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
