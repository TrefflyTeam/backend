package eventdto

import (
	"github.com/jackc/pgx/v5/pgtype"
	"time"
	db "treffly/db/sqlc"
)

type EventResponse struct {
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

type EventsListResponse struct {
	Events []EventResponse `json:"Events"`
}

type CreateEventResponse struct {
	EventResponse
	ImageEventURL string        `json:"image_event_url"`
	ImageUserURL  string        `json:"image_user_url"`
}

func NewCreateEventResponse(event EventResponse, imageEventURL, imageUserURL string) CreateEventResponse {
	return CreateEventResponse{
		EventResponse: event,
		ImageEventURL: imageEventURL,
		ImageUserURL:  imageUserURL,
	}
}

func NewEventsListResponse(Events []db.EventRow) EventsListResponse {
	return EventsListResponse{Events: ConvertEvents(Events)}
}

type EventByIDResponse struct {
	EventResponse
	IsOwner       bool `json:"is_owner"`
	IsParticipant bool `json:"is_participant"`
}

func NewEventByIDResponse(event EventResponse, isOwner, isParticipant bool) EventByIDResponse {
	return EventByIDResponse{
		EventResponse: event,
		IsOwner:       isOwner,
		IsParticipant: isParticipant,
	}
}

type GetHomeEventsResponse struct {
	Premium     []EventResponse `json:"premium"`
	Recommended []EventResponse `json:"recommended"`
	Latest      []EventResponse `json:"latest"`
	Popular     []EventResponse `json:"popular"`
}

func NewGetHomeEventsResponse(
	premium []db.EventRow,
	recommended []db.EventRow,
	latest []db.EventRow,
	popular []db.EventRow,
) GetHomeEventsResponse {
	return GetHomeEventsResponse{
		Premium:     ConvertEvents(premium),
		Recommended: ConvertEvents(recommended),
		Latest:      ConvertEvents(latest),
		Popular:     ConvertEvents(popular),
	}
}
