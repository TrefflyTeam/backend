package api

import (
	db "treffly/db/sqlc"
)

func convertEvents(events []db.EventResponse) []eventResponse {
	result := make([]eventResponse, 0, len(events))
	for _, e := range events {
		result = append(result, eventResponse{
			ID:          e.ID,
			Name:        e.Name,
			Description: e.Description,
			Date:        e.Date,
			Address:     e.Address,
			OwnerUsername: e.OwnerUsername,
			Latitude:    e.Latitude,
			Longitude:   e.Longitude,
			IsPrivate:   e.IsPrivate,
			IsPremium:   e.IsPremium,
			CreatedAt:   e.CreatedAt,
			Tags: 		 e.Tags,
			ParticipantCount: int32(e.ParticipantsCount),
		})
	}
	return result
}

