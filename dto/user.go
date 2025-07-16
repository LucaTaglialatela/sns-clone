package dto

import "github.com/HENNGE/snsclone-202506-golang-luca/entity"

type UserProfileResponse struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Following []string `json:"following"`
}

type CreateUserRequest struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Picture string `json:"picture"`
}

type FollowRequest struct {
	FollowingID string `json:"following_id"`
}

type UnfollowRequest struct {
	UnfollowingID string `json:"unfollowing_id"`
}

type User struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Picture string `json:"picture"`
}

type Follow struct {
	FollowingID string `json:"following_id"`
}

type Following struct {
	FollowingIDs []string `json:"following_ids"`
}

func (u *User) FromEntity(user *entity.User) {
	u.ID = user.ID
	u.Name = user.Name
	u.Email = user.Email
	u.Picture = user.Picture
}

func (f *Follow) FromEntity(follow *entity.Follow) {
	f.FollowingID = follow.FollowedID
}

func (f *Following) FromEntity(following []*entity.Follow) {
	followingIDs := make([]string, 0, len(following))
	for _, follow := range following {
		followingIDs = append(followingIDs, follow.FollowedID)
	}
	f.FollowingIDs = followingIDs
}
