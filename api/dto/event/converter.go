package eventdto

import (
	"treffly/api/common"
	eventservice "treffly/api/service/event"
)

type EventConverter struct {
	env    string
	domain string
}

func NewEventConverter(env, domain string) *EventConverter {
	return &EventConverter{
		env:    env,
		domain: domain,
	}
}

func (c *EventConverter) ToEventWithParticipantsResponse(e eventservice.EventWithParticipants) EventWithParticipantsResponse {
	return EventWithParticipantsResponse{
		EventWithTagsResponse: c.ToEventWithTagsResponse(e.EventWithTags),
		ParticipantCount:      e.ParticipantCount,
	}
}

func (c *EventConverter) ConvertEventToResponse(e eventservice.Event) EventResponse {
	return EventResponse{
		ID:          e.ID,
		Name:        e.Name,
		Description: e.Description,
		Capacity:    e.Capacity,
		Latitude:    e.Latitude,
		Longitude:   e.Longitude,
		Address:     e.Address,
		Date:        e.Date,
		IsPrivate:   e.IsPrivate,
		IsPremium:   e.IsPremium,
		CreatedAt:   e.CreatedAt,
	}
}

func (c *EventConverter) ConvertEventWithOwnerToResponse(e eventservice.EventWithOwner) EventWithOwnerResponse {
	return EventWithOwnerResponse{
		EventResponse: c.ConvertEventToResponse(e.Event),
		OwnerUsername: e.OwnerUsername,
	}
}

func (c *EventConverter) ToEventWithTagsResponse(e eventservice.EventWithTags) EventWithTagsResponse {
	return EventWithTagsResponse{
		EventWithOwnerResponse: c.ConvertEventWithOwnerToResponse(e.EventWithOwner),
		Tags:                   c.convertTagsToResponse(e.Tags),
	}
}

func (c *EventConverter) ToEventWithImagesResponse(e *eventservice.EventWithImages) EventWithImagesResponse {
	return EventWithImagesResponse{
		EventWithParticipantsResponse: c.ToEventWithParticipantsResponse(e.EventWithParticipants),
		ImageEventURL:                 common.ImageURL(c.env, c.domain, e.ImageEventPath),
		ImageUserURL:                  common.ImageURL(c.env, c.domain, e.ImageUserPath),
	}
}

func (c *EventConverter) ToEventWithMetaResponse(e *eventservice.EventWithMeta) EventWithMetaResponse {
	return EventWithMetaResponse{
		EventWithImagesResponse: c.ToEventWithImagesResponse(&e.EventWithImages),
		IsOwner:                 e.IsOwner,
		IsParticipant:           e.IsParticipant,
	}
}

func (c *EventConverter) convertTagsToResponse(tags []eventservice.Tag) []TagResponse {
	result := make([]TagResponse, len(tags))
	for i, t := range tags {
		result[i] = TagResponse{
			ID:   t.ID,
			Name: t.Name,
		}
	}
	return result
}

func (c *EventConverter) ToEventsWithImages(events []eventservice.EventWithImages) []EventWithImagesResponse {
	result := make([]EventWithImagesResponse, len(events))
	for i, e := range events {
		result[i] = c.ToEventWithImagesResponse(&e)
	}
	return result
}

func (c *EventConverter) ToHomeEventsResponse(h *eventservice.HomeEvents) HomeEventsResponse {
	return HomeEventsResponse{
		Premium:     c.ToEventsWithImages(h.Premium),
		Recommended: c.ToEventsWithImages(h.Recommended),
		Latest:      c.ToEventsWithImages(h.Latest),
		Popular:     c.ToEventsWithImages(h.Popular),
	}
}

