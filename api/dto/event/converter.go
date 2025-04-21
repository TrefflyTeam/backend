package eventdto

import (
	db "treffly/db/sqlc"
)

func ConvertEvent(event db.EventRow) Event {
	result := Event{
		ID:               event.GetID(),
		Name:             event.GetName(),
		Description:      event.GetDescription(),
		Date:             event.GetDate(),
		Address:          event.GetAddress(),
		Capacity:         event.GetCapacity(),
		OwnerUsername:    event.GetOwnerUsername(),
		Latitude:         event.GetLatitude(),
		Longitude:        event.GetLongitude(),
		IsPrivate:        event.GetIsPrivate(),
		IsPremium:        event.GetIsPremium(),
		CreatedAt:        event.GetCreatedAt(),
		Tags:             event.GetTags(),
		ParticipantCount: int32(event.GetParticipantsCount()),
	}

	return result
}

func ConvertEvents(events []db.EventRow) []Event {
	result := make([]Event, 0, len(events))
	for _, e := range events {
		result = append(result, ConvertEvent(e))
	}

	return result
}
