package models

import (
	"github.com/google/uuid"
	"time"
)

type User struct {
	ID           int32
	Username     string
	Email        string
	CreatedAt    time.Time
}

type UserWithTags struct {
	User
	Tags      []Tag
	ImagePath string
}

type CreateUserParams struct {
	Username string
	Email    string
	Password string
}

type LoginParams struct {
	Email    string
	Password string
}

type RefreshSessionParams struct {
	UUID         uuid.UUID
	UserID       int32
	RefreshToken string
	ExpiresAt    time.Time
}

type UpdateUserParams struct {
	ID          int32
	Username    string
	NewImageID  uuid.UUID
	Path        string
	OldImageID  uuid.UUID
	DeleteImage bool
}

type UpdateUserTagsParams struct {
	UserID int32
	TagIDs []int32
}

