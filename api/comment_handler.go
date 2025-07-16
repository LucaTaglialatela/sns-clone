package api

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"

	"github.com/HENNGE/snsclone-202506-golang-luca/dto"
	"github.com/HENNGE/snsclone-202506-golang-luca/service"
)

type CommentHandler struct {
	service service.DefaultCommentService
}

func NewCommentHandler(service service.DefaultCommentService) *CommentHandler {
	return &CommentHandler{
		service: service,
	}
}

func (h *CommentHandler) Create(w http.ResponseWriter, r *http.Request) {
	postId := r.PathValue("post_id")
	if postId == "" {
		log.Printf("post id is empty")
		http.Error(w, "post id cannot be empty", http.StatusBadRequest)
		return
	}

	claims, ok := r.Context().Value(userClaimsKey).(*AppClaims)
	if !ok {
		http.Error(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	request := dto.SaveCommentRequest{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Remove invisible characters from request text
	var re = regexp.MustCompile("[\u0000-\u001F\u00A0\u115F\u1160\u2000-\u200D\u2028-\u202F\u205F\u2060\u3000\u3164\uFEFF\n\r]")
	request.Text = re.ReplaceAllString(request.Text, "")

	if request.Text == "" {
		http.Error(w, "comment cannot be empty", http.StatusBadRequest)
		return
	}

	if len(request.Text) > 140 {
		http.Error(w, "comment length exceeds the maximum", http.StatusBadRequest)
		return
	}

	comment, err := h.service.Create(r.Context(), postId, claims.UserID, claims.UserName, &request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(comment)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *CommentHandler) GetByPostID(w http.ResponseWriter, r *http.Request) {
	postId := r.PathValue("post_id")

	comments, err := h.service.GetByPostID(r.Context(), postId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(comments)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *CommentHandler) Update(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("user_id")
	postId := r.PathValue("post_id")
	commentId := r.PathValue("comment_id")

	isAuthorized := IsAuthorized(r.Context(), userId)
	if !isAuthorized {
		http.Error(w, "Not authorized", http.StatusUnauthorized)
		return
	}

	request := dto.SaveCommentRequest{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Remove invisible characters from request text
	var re = regexp.MustCompile(`[\x{200B}-\x{200D}\x{FEFF}\x{2060}-\x{206F}\n\r]`)
	request.Text = re.ReplaceAllString(request.Text, "")

	if request.Text == "" {
		http.Error(w, "comment cannot be empty", http.StatusBadRequest)
		return
	}

	if len(request.Text) > 140 {
		http.Error(w, "comment length exceeds the maximum", http.StatusBadRequest)
		return
	}

	updatedComment, err := h.service.Update(r.Context(), postId, commentId, &request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(updatedComment)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *CommentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("user_id")
	postId := r.PathValue("post_id")
	commentId := r.PathValue("comment_id")

	isAuthorized := IsAuthorized(r.Context(), userId)
	if !isAuthorized {
		http.Error(w, "Not authorized", http.StatusUnauthorized)
		return
	}

	err := h.service.Delete(r.Context(), postId, commentId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
