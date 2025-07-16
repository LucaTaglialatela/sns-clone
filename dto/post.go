package dto

import (
	"time"

	"github.com/HENNGE/snsclone-202506-golang-luca/entity"
)

type CreatePostRequest struct {
	Text  string `json:"text"`
	Image string `json:"image,omitempty"`
}

type UpdatePostRequest struct {
	Text  string `json:"text"`
	Image string `json:"image,omitempty"`
}

type Post struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	UserName  string     `json:"user_name"`
	Text      string     `json:"text"`
	Image     string     `json:"image,omitempty"`
	Timestamp time.Time  `json:"timestamp"`
	ImageURL  string     `json:"image_url,omitempty"`
	Edited    *time.Time `json:"edited"`
}

func (p *Post) FromEntity(post *entity.Post) {
	p.ID = post.ID
	p.UserID = post.UserID
	p.UserName = post.UserName
	p.Text = post.Text
	p.Image = post.Image
	p.Timestamp = post.Timestamp

	if post.Edited != nil {
		p.Edited = post.Edited
	}

	if post.ImageURL != nil {
		p.ImageURL = *post.ImageURL
	}
}
