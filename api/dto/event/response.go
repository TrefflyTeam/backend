package eventdto

import (
	"time"
)

type EventResponse struct {
	ID               int32         `json:"id"`
	Name             string        `json:"name"`
	Description      string        `json:"description,omitempty"`
	Capacity         int32         `json:"capacity"`
	Latitude         float64       `json:"latitude"`
	Longitude        float64       `json:"longitude"`
	Address          string        `json:"address"`
	Date             time.Time     `json:"date"`
	IsPrivate        bool          `json:"is_private"`
	IsPremium        bool          `json:"is_premium"`
	CreatedAt        time.Time     `json:"created_at"`
	OwnerUsername    string        `json:"owner_username"`
	IsOwner          bool          `json:"is_owner"`
	IsParticipant    bool          `json:"is_participant"`
	Tags             []TagResponse `json:"tags"`
	ParticipantCount int           `json:"participant_count"`
	ImageEventURL    string        `json:"image_event_url"`
	ImageUserURL     string        `json:"image_user_url"`
}

type TagResponse struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}

type HomeEventsResponse struct {
	Premium     []EventResponse `json:"premium"`
	Recommended []EventResponse `json:"recommended"`
	Latest      []EventResponse `json:"latest"`
	Popular     []EventResponse `json:"popular"`
}
