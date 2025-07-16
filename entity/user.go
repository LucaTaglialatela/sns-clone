package entity

import "fmt"

type User struct {
	PK      string `dynamodbav:"pk"`
	SK      string `dynamodbav:"sk"`
	ID      string `dynamodbav:"id"`
	Name    string `dynamodbav:"name"`
	Email   string `dynamodbav:"email"`
	Picture string `dynamodbav:"picture"`
	// Image
}

type Follow struct {
	PK         string `dynamodbav:"pk"`
	SK         string `dynamodbav:"sk"`
	FollowedID string `dynamodbav:"id"`
}

type Unfollow struct {
	PK string `dynamodbav:"pk"`
	SK string `dynamodbav:"sk"`
}

func NewFollow(followerId, followingId string) (*Follow, error) {
	f := &Follow{
		PK:         fmt.Sprintf("user#%s", followerId),
		SK:         fmt.Sprintf("follower#%s", followingId),
		FollowedID: followingId,
	}
	return f, nil
}

func NewUnfollow(followerId, followingId string) (*Unfollow, error) {
	f := &Unfollow{
		PK: fmt.Sprintf("user#%s", followerId),
		SK: fmt.Sprintf("follower#%s", followingId),
	}
	return f, nil
}

func NewUser(id, name, email, picture string) (*User, error) {
	u := &User{
		PK:      "user",
		SK:      id,
		ID:      id,
		Name:    name,
		Email:   email,
		Picture: picture,
	}
	return u, nil
}
