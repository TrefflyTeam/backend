package api

import (
	db "treffly/db/sqlc"
)

func convertEvent(event db.EventRow) eventResponse {
	result := eventResponse{
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

func convertEvents(events []db.EventRow) []eventResponse {
	result := make([]eventResponse, 0, len(events))
	for _, e := range events {
		result = append(result, convertEvent(e))
	}

	return result
}
