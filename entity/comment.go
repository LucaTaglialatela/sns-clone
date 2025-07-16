package entity

import (
	"fmt"
	"time"

	"github.com/oklog/ulid/v2"
)

type Comment struct {
	PK        string     `dynamodbav:"pk"`
	SK        string     `dynamodbav:"sk"`
	ID        string     `dynamodbav:"id"`
	UserID    string     `dynamodbav:"user_id"`
	UserName  string     `dynamodbav:"name"`
	Text      string     `dynamodbav:"text"`
	Timestamp time.Time  `dynamodbav:"timestamp"`
	Edited    *time.Time `dynamodbav:"edited"`
}

func NewComment(postId, userId, userName, text string) (*Comment, error) {
	ulid := ulid.Make().String()
	p := &Comment{
		PK:        fmt.Sprintf("post#%s", postId),
		SK:        fmt.Sprintf("comment#%s", ulid),
		ID:        ulid,
		UserID:    userId,
		UserName:  userName,
		Text:      text,
		Timestamp: time.Now(),
		Edited:    nil,
	}

	return p, nil
}
