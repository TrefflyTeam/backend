package eventdto

import (
	"github.com/jackc/pgx/v5/pgtype"
	"time"
	db "treffly/db/sqlc"
)

type Event struct {
	ID               int32          `json:"id"`
	Name             string         `json:"name"`
	Description      string         `json:"description"`
	Capacity         int32          `json:"capacity"`
	Latitude         pgtype.Numeric `json:"latitude"`
	Longitude        pgtype.Numeric `json:"longitude"`
	Address          string         `json:"address"`
	Date             time.Time      `json:"date"`
	IsPrivate        bool           `json:"is_private"`
	IsPremium        bool           `json:"is_premium"`
	CreatedAt        time.Time      `json:"created_at"`
	OwnerUsername    string         `json:"owner_username"`
	Tags             []db.Tag       `json:"tags"`
	ParticipantCount int32          `json:"participant_count"`
}

type EventsList struct {
	Events []Event `json:"Events"`
}

func NewEventsList(Events []db.EventRow) EventsList {
	return EventsList{Events: ConvertEvents(Events)}
}

type EventByID struct{
	Event
	IsOwner bool `json:"is_owner"`
	IsParticipant bool `json:"is_participant"`
}

func NewEventByID(event Event, isOwner, isParticipant bool) EventByID {
	return EventByID{
		Event: event,
		IsOwner: isOwner,
		IsParticipant: isParticipant,
	}
}

type GetHomeEvents struct {
	Premium     []Event `json:"premium"`
	Recommended []Event `json:"recommended"`
	Latest      []Event `json:"latest"`
	Popular     []Event `json:"popular"`
}

func NewGetHomeEvents(
	premium []db.EventRow,
	recommended []db.EventRow,
	latest []db.EventRow,
	popular []db.EventRow,
) GetHomeEvents {
	return GetHomeEvents{
		Premium:     ConvertEvents(premium),
		Recommended: ConvertEvents(recommended),
		Latest:      ConvertEvents(latest),
		Popular:     ConvertEvents(popular),
	}
}
