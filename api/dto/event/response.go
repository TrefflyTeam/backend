package eventdto

import (
	"github.com/jackc/pgx/v5/pgtype"
	"time"
)

type EventResponse struct {
	ID          int32          `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Capacity    int32          `json:"capacity"`
	Latitude    pgtype.Numeric `json:"latitude"`
	Longitude   pgtype.Numeric `json:"longitude"`
	Address     string         `json:"address"`
	Date        time.Time      `json:"date"`
	IsPrivate   bool           `json:"is_private"`
	IsPremium   bool           `json:"is_premium"`
	CreatedAt   time.Time      `json:"created_at"`
}

type EventWithOwnerResponse struct {
	EventResponse
	OwnerUsername string `json:"owner_username"`
}

type EventWithTagsResponse struct {
	EventWithOwnerResponse
	Tags []TagResponse `json:"tags"`
}

type EventWithParticipantsResponse struct {
	EventWithTagsResponse
	ParticipantCount int32 `json:"participant_count"`
}

type EventWithImagesResponse struct {
	EventWithParticipantsResponse
	ImageEventURL string `json:"image_event_url"`
	ImageUserURL  string `json:"image_user_url"`
}

type EventWithMetaResponse struct {
	EventWithImagesResponse
	IsOwner       bool `json:"is_owner"`
	IsParticipant bool `json:"is_participant"`
}

type TagResponse struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}

type HomeEventsResponse struct {
	Premium     []EventWithImagesResponse `json:"premium"`
	Recommended []EventWithImagesResponse `json:"recommended"`
	Latest      []EventWithImagesResponse `json:"latest"`
	Popular     []EventWithImagesResponse `json:"popular"`
}
