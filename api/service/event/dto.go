package eventservice

import (
	"github.com/jackc/pgx/v5/pgtype"
	"time"
	db "treffly/db/sqlc"
)

type EventWithStatus struct {
	Event db.EventRow
	IsParticipant bool
	IsOwner       bool
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

type HomeEvents struct {
	Premium     []db.EventRow
	Recommended []db.EventRow
	Latest      []db.EventRow
	Popular     []db.EventRow
}

type SubscriptionParams struct {
	EventID int32
	UserID  int32
}

type UserEventsParams struct {
	UserID int32
}