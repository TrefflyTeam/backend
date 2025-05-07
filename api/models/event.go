package models

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"time"
)

type Event struct {
	ID               int32
	Name             string
	Description      string
	Capacity         int32
	Latitude         float64
	Longitude        float64
	Address          string
	Date             time.Time
	IsPrivate        bool
	IsPremium        bool
	CreatedAt        time.Time
	OwnerUsername    string
	IsOwner          bool
	IsParticipant     bool
	Tags             []Tag
	ParticipantCount int
	ImagePath        string
	OwnerImagePath   string
}

type Tag struct {
	ID   int32
	Name string
}

type HomeEvents struct {
	Premium     []Event
	Recommended []Event
	Latest      []Event
	Popular     []Event
}

type CreateParams struct {
	Name        string
	Description string
	Capacity    int32
	Latitude    pgtype.Numeric
	Longitude   pgtype.Numeric
	Address     string
	Date        time.Time
	IsPrivate   bool
	Tags        []int32
	OwnerID     int32
	ImageID     uuid.UUID
	Path        string
}

type ListParams struct {
	Lat       pgtype.Numeric
	Lon       pgtype.Numeric
	Search    string
	TagIDs    []int32
	DateRange string
}

type UpdateParams struct {
	EventID     int32
	Name        string
	Description string
	Capacity    int32
	Latitude    pgtype.Numeric
	Longitude   pgtype.Numeric
	Address     string
	Date        time.Time
	IsPrivate   bool
	Tags        []int32
	UserID      int32
	NewImageID  uuid.UUID
	Path        string
	DeleteImage bool
	OldImageID  uuid.UUID
}

type DeleteParams struct {
	EventID int32
	UserID  int32
}

type GetHomeParams struct {
	UserID int32
	Lat    pgtype.Numeric
	Lon    pgtype.Numeric
}

type SubscriptionParams struct {
	EventID int32
	UserID  int32
}

type UserEventsParams struct {
	UserID int32
}
