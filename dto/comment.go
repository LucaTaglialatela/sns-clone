package dto

import (
	"time"

	"github.com/HENNGE/snsclone-202506-golang-luca/entity"
)

type SaveCommentRequest struct {
	Text string `json:"text"`
}

type Comment struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	UserName  string     `json:"user_name"`
	Text      string     `json:"text"`
	Timestamp time.Time  `json:"timestamp"`
	Edited    *time.Time `json:"edited"`
}

func (c *Comment) FromEntity(comment *entity.Comment) {
	c.ID = comment.ID
	c.UserID = comment.UserID
	c.UserName = comment.UserName
	c.Text = comment.Text
	c.Timestamp = comment.Timestamp

	if comment.Edited != nil {
		c.Edited = comment.Edited
	}
}
