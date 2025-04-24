package userdto

import (
	"time"
	db "treffly/db/sqlc"
)

type UserResponse struct {
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

func NewUserResponse(user db.User) UserResponse {
	return UserResponse{
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}
}

type UserWithTagsResponse struct {
	ID        int32     `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Tags      []db.Tag  `json:"tags"`
	CreatedAt time.Time `json:"created_at"`
}

func NewUserWithTagsResponse(user db.UserWithTagsView) UserWithTagsResponse {
	return UserWithTagsResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Tags:      user.Tags,
		CreatedAt: user.CreatedAt,
	}
}

type LoginResponse struct {
	UserResponse
}

func NewLoginResponse(user db.User) LoginResponse {
	return LoginResponse{
		UserResponse: NewUserResponse(user),
	}
}

type UpdateUserResponse struct {
	UserWithTagsResponse
	ImageURL string `json:"image_url"`
}

func NewUpdateUserResponse(user db.UserWithTagsView, imageURL string) UpdateUserResponse {
	userWithTags := NewUserWithTagsResponse(user)
	return UpdateUserResponse{
		UserWithTagsResponse: userWithTags,
		ImageURL: imageURL,
	}
}
