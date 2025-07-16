package entity

import (
	"fmt"
	"time"

	"github.com/oklog/ulid/v2"
)

type Post struct {
	PK        string     `dynamodbav:"pk"`
	SK        string     `dynamodbav:"sk"`
	GSIPK     string     `dynamodbav:"gsi1_pk"`
	GSISK     string     `dynamodbav:"gsi1_sk"`
	ID        string     `dynamodbav:"id"`
	UserID    string     `dynamodbav:"user_id"`
	UserName  string     `dynamodbav:"name"`
	Text      string     `dynamodbav:"text"`
	Timestamp time.Time  `dynamodbav:"timestamp"`
	Image     string     `dynamodbav:"image"`
	Edited    *time.Time `dynamodbav:"edited"`
	ImageURL  *string
}

func NewPost(userId, userName, text, image string) (*Post, error) {
	ulid := ulid.Make().String()
	p := &Post{
		PK:        fmt.Sprintf("user#%s", userId),
		SK:        fmt.Sprintf("post#%s", ulid),
		GSIPK:     "timeline",
		GSISK:     ulid,
		ID:        ulid,
		UserID:    userId,
		UserName:  userName,
		Text:      text,
		Timestamp: time.Now(),
		Image:     image,
		Edited:    nil,
	}

	return p, nil
}
