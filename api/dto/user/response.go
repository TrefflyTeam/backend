package userdto

import (
	"time"
)

type UserResponse struct {
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type UserWithTagsResponse struct {
	UserResponse
	Tags     []TagResponse `json:"tags"`
	ImageURL string        `json:"image_url"`
}

type TagResponse struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}

type AdminUserResponse struct {
	ID int32 `json:"id"`
	UserResponse
}
