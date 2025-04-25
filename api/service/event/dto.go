package eventservice

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"time"
)

type Event struct {
	ID          int32
	Name        string
	Description string
	Capacity    int32
	Latitude    pgtype.Numeric
	Longitude   pgtype.Numeric
	Address     string
	Date        time.Time
	IsPrivate   bool
	IsPremium   bool
	CreatedAt   time.Time
}

type EventWithOwner struct {
	Event
	OwnerUsername string
}

type EventWithTags struct {
	EventWithOwner
	Tags []Tag
}

type EventWithParticipants struct {
	EventWithTags
	ParticipantCount int32
}

type EventWithImages struct {
	EventWithParticipants
	ImageEventPath string
	ImageUserPath  string
}

type EventWithMeta struct {
	EventWithImages
	IsOwner       bool
	IsParticipant bool
}

type Tag struct {
	ID   int32
	Name string
}

type HomeEvents struct {
	Premium     []EventWithImages
	Recommended []EventWithImages
	Latest      []EventWithImages
	Popular     []EventWithImages
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
	ImageID     uuid.UUID
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
