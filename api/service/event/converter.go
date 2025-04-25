package eventservice

import (
	"github.com/jackc/pgx/v5/pgtype"
	db "treffly/db/sqlc"
)

func convertTags(dbTags []db.Tag) []Tag {
	tags := make([]Tag, len(dbTags))
	for i, t := range dbTags {
		tags[i] = Tag{
			ID:   t.ID,
			Name: t.Name,
		}
	}
	return tags
}

func ConvertGetEventRow(e db.GetEventRow, isOwner, isParticipant bool) EventWithMeta {
	base := Event{
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

	return EventWithMeta{
		EventWithImages: EventWithImages{
			EventWithParticipants: EventWithParticipants{
				EventWithTags: EventWithTags{
					EventWithOwner: EventWithOwner{
						Event:         base,
						OwnerUsername: safeString(e.OwnerUsername),
					},
					Tags: convertTags(e.Tags),
				},
				ParticipantCount: int32(e.ParticipantsCount),
			},
			ImageEventPath: safeString(e.EventImagePath),
			ImageUserPath:  safeString(e.UserImagePath),
		},
		IsOwner:       isOwner,
		IsParticipant: isParticipant,
	}
}

func ConvertListEventsRow(e db.ListEventsRow) EventWithImages {
	base := Event{
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

	return EventWithImages{
		EventWithParticipants: EventWithParticipants{
			EventWithTags: EventWithTags{
				EventWithOwner: EventWithOwner{
					Event:         base,
					OwnerUsername: safeString(e.OwnerUsername),
				},
				Tags: convertTags(e.Tags),
			},
			ParticipantCount: int32(e.ParticipantsCount),
		},
		ImageEventPath: safeString(e.EventImagePath),
	}
}

func safeString(s pgtype.Text) string {
	if s.Valid {
		return s.String
	}
	return ""
}

type RecommendedRow interface {
	db.GetUserRecommendedEventsRow | db.GetGuestRecommendedEventsRow
}

func ConvertHomeEvents[T RecommendedRow](
	premium []db.GetPremiumEventsRow,
	recommended []T,
	latest []db.GetLatestEventsRow,
	popular []db.GetPopularEventsRow,
) HomeEvents {
	return HomeEvents{
		Premium:     convertEventType(premium),
		Recommended: convertRecommendedEvents(recommended),
		Latest:      convertEventType(latest),
		Popular:     convertEventType(popular),
	}
}

func convertRecommendedEvents[T RecommendedRow](rows []T) []EventWithImages {
	result := make([]EventWithImages, len(rows))
	for i, row := range rows {
		result[i] = convertSingleRecommended(row)
	}
	return result
}

func convertSingleRecommended[T RecommendedRow](row T) EventWithImages {
	switch v := any(row).(type) {
	case db.GetUserRecommendedEventsRow:
		return convertUserRecommended(v)
	case db.GetGuestRecommendedEventsRow:
		return convertGuestRecommended(v)
	default:
		panic("unsupported type")
	}
}

func convertUserRecommended(e db.GetUserRecommendedEventsRow) EventWithImages {
	base := Event{
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

	return EventWithImages{
		EventWithParticipants: EventWithParticipants{
			EventWithTags: EventWithTags{
				EventWithOwner: EventWithOwner{
					Event:         base,
					OwnerUsername: safeString(e.OwnerUsername),
				},
				Tags: convertTags(e.Tags),
			},
			ParticipantCount: int32(e.ParticipantsCount),
		},
		ImageEventPath: safeString(e.EventImagePath),
	}
}

func convertGuestRecommended(e db.GetGuestRecommendedEventsRow) EventWithImages {
	base := Event{
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

	return EventWithImages{
		EventWithParticipants: EventWithParticipants{
			EventWithTags: EventWithTags{
				EventWithOwner: EventWithOwner{
					Event:         base,
					OwnerUsername: safeString(e.OwnerUsername),
				},
				Tags: convertTags(e.Tags),
			},
			ParticipantCount: int32(e.ParticipantsCount),
		},
		ImageEventPath: safeString(e.EventImagePath),
	}
}

func convertEventType[T any](rows []T) []EventWithImages {
	result := make([]EventWithImages, len(rows))
	for i, row := range rows {
		switch v := any(row).(type) {
		case db.ListEventsRow:
			result[i] = ConvertListEventsRow(v)
		case db.GetPremiumEventsRow:
			result[i] = convertPremiumEvent(v)
		case db.GetGuestRecommendedEventsRow:
			result[i] = convertSingleRecommended(v)
		case db.GetUserRecommendedEventsRow:
			result[i] = convertSingleRecommended(v)
		case db.GetLatestEventsRow:
			result[i] = convertLatestEvent(v)
		case db.GetPopularEventsRow:
			result[i] = convertPopularEvent(v)
		case db.GetUpcomingUserEventsRow:
			result[i] = convertUpcomingEventsRow(v)
		case db.GetPastUserEventsRow:
			result[i] = convertPastEventsRow(v)
		case db.GetOwnedUserEventsRow:
			result[i] = convertOwnedEventsRow(v)
		}
	}
	return result
}

func convertOwnedEventsRow(e db.GetOwnedUserEventsRow) EventWithImages {
	base := Event{
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

	return EventWithImages{
		EventWithParticipants: EventWithParticipants{
			EventWithTags: EventWithTags{
				EventWithOwner: EventWithOwner{
					Event:         base,
					OwnerUsername: safeString(e.OwnerUsername),
				},
				Tags: convertTags(e.Tags),
			},
			ParticipantCount: int32(e.ParticipantsCount),
		},
		ImageEventPath: safeString(e.EventImagePath),
	}
}

func convertPastEventsRow(e db.GetPastUserEventsRow) EventWithImages {
	base := Event{
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

	return EventWithImages{
		EventWithParticipants: EventWithParticipants{
			EventWithTags: EventWithTags{
				EventWithOwner: EventWithOwner{
					Event:         base,
					OwnerUsername: safeString(e.OwnerUsername),
				},
				Tags: convertTags(e.Tags),
			},
			ParticipantCount: int32(e.ParticipantsCount),
		},
		ImageEventPath: safeString(e.EventImagePath),
	}
}

func convertUpcomingEventsRow(e db.GetUpcomingUserEventsRow) EventWithImages {
	base := Event{
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

	return EventWithImages{
		EventWithParticipants: EventWithParticipants{
			EventWithTags: EventWithTags{
				EventWithOwner: EventWithOwner{
					Event:         base,
					OwnerUsername: safeString(e.OwnerUsername),
				},
				Tags: convertTags(e.Tags),
			},
			ParticipantCount: int32(e.ParticipantsCount),
		},
		ImageEventPath: safeString(e.EventImagePath),
	}
}

func convertPremiumEvent(e db.GetPremiumEventsRow) EventWithImages {
	base := Event{
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

	return EventWithImages{
		EventWithParticipants: EventWithParticipants{
			EventWithTags: EventWithTags{
				EventWithOwner: EventWithOwner{
					Event:         base,
					OwnerUsername: safeString(e.OwnerUsername),
				},
				Tags: convertTags(e.Tags),
			},
			ParticipantCount: int32(e.ParticipantsCount),
		},
		ImageEventPath: safeString(e.EventImagePath),
	}
}

func convertLatestEvent(e db.GetLatestEventsRow) EventWithImages {
	base := Event{
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

	return EventWithImages{
		EventWithParticipants: EventWithParticipants{
			EventWithTags: EventWithTags{
				EventWithOwner: EventWithOwner{
					Event:         base,
					OwnerUsername: safeString(e.OwnerUsername),
				},
				Tags: convertTags(e.Tags),
			},
			ParticipantCount: int32(e.ParticipantsCount),
		},
		ImageEventPath: safeString(e.EventImagePath),
	}
}

func convertPopularEvent(e db.GetPopularEventsRow) EventWithImages {
	base := Event{
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

	return EventWithImages{
		EventWithParticipants: EventWithParticipants{
			EventWithTags: EventWithTags{
				EventWithOwner: EventWithOwner{
					Event:         base,
					OwnerUsername: safeString(e.OwnerUsername),
				},
				Tags: convertTags(e.Tags),
			},
			ParticipantCount: int32(e.ParticipantsCount),
		},
		ImageEventPath: safeString(e.EventImagePath),
	}
}
