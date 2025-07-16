package api

import (
	"encoding/json"
	"net/http"
	"regexp"

	"github.com/HENNGE/snsclone-202506-golang-luca/dto"
	"github.com/HENNGE/snsclone-202506-golang-luca/entity"
	"github.com/HENNGE/snsclone-202506-golang-luca/service"
)

type PostHandler struct {
	Service service.DefaultPostService
	Broker  *entity.Broker
}

func NewPostHandler(service service.DefaultPostService, broker *entity.Broker) *PostHandler {
	return &PostHandler{
		Service: service,
		Broker:  broker,
	}
}

func (h *PostHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(userClaimsKey).(*AppClaims)
	if !ok {
		http.Error(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	request := dto.CreatePostRequest{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Remove invisible characters from request text
	var re = regexp.MustCompile("[\u0000-\u0009\u000B-\u000C\u000E-\u001F\u00A0\u115F\u1160\u2000-\u200D\u202A-\u202F\u205F\u2060\u3000\u3164\uFEFF]")
	request.Text = re.ReplaceAllString(request.Text, "")

	if request.Text == "" && request.Image == "" {
		http.Error(w, "post cannot be empty", http.StatusBadRequest)
		return
	}

	if len(request.Text) > 280 {
		http.Error(w, "post length exceeds the maximum", http.StatusBadRequest)
		return
	}

	post, err := h.Service.Create(r.Context(), claims.UserID, claims.UserName, &request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	eventData, err := json.Marshal(post)
	if err != nil {
		http.Error(w, "Failed to create post", http.StatusInternalServerError)
		return
	}
	event := entity.SSEEvent{
		Name: "new_post",
		Data: string(eventData),
	}
	h.Broker.Messages <- event

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(post)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *PostHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	posts, err := h.Service.GetAll(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(posts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *PostHandler) GetByUserID(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("user_id")

	posts, err := h.Service.GetByUserID(r.Context(), userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(posts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *PostHandler) Update(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("user_id")
	postId := r.PathValue("post_id")

	isAuthorized := IsAuthorized(r.Context(), userId)
	if !isAuthorized {
		http.Error(w, "Not authorized", http.StatusUnauthorized)
		return
	}

	request := dto.UpdatePostRequest{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Remove invisible characters from request text
	var re = regexp.MustCompile(`[\x{200B}-\x{200D}\x{FEFF}\x{2060}-\x{206F}]`)
	request.Text = re.ReplaceAllString(request.Text, "")

	if request.Text == "" && request.Image == "" {
		http.Error(w, "post cannot be empty", http.StatusBadRequest)
		return
	}

	if len(request.Text) > 280 {
		http.Error(w, "post length exceeds the maximum", http.StatusBadRequest)
		return
	}

	updatedPost, err := h.Service.Update(r.Context(), userId, postId, &request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	eventData, err := json.Marshal(updatedPost)
	if err != nil {
		http.Error(w, "Failed to create post", http.StatusInternalServerError)
		return
	}
	event := entity.SSEEvent{
		Name: "update_post",
		Data: string(eventData),
	}
	h.Broker.Messages <- event

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(updatedPost)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *PostHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("user_id")
	postId := r.PathValue("post_id")

	isAuthorized := IsAuthorized(r.Context(), userId)
	if !isAuthorized {
		http.Error(w, "Not authorized", http.StatusUnauthorized)
		return
	}

	err := h.Service.Delete(r.Context(), userId, postId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	eventData, err := json.Marshal(map[string]string{"id": postId})
	if err != nil {
		http.Error(w, "Failed to create post", http.StatusInternalServerError)
		return
	}
	event := entity.SSEEvent{
		Name: "delete_post",
		Data: string(eventData),
	}
	h.Broker.Messages <- event

	w.WriteHeader(http.StatusNoContent)
}
