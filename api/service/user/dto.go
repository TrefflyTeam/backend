package userservice

import (
	"github.com/google/uuid"
	"time"
)

type CreateParams struct {
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
	ID       int32
	Username string
}

type UpdateUserTagsParams struct {
	UserID int32
	TagIDs []int32
}
