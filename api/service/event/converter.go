package eventservice

import (
	"github.com/jackc/pgx/v5/pgtype"
	"treffly/api/models"
	db "treffly/db/sqlc"
	"treffly/util"
)

func convertTags(dbTags []db.Tag) []models.Tag {
	tags := make([]models.Tag, len(dbTags))
	for i, t := range dbTags {
		tags[i] = models.Tag{
			ID:   t.ID,
			Name: t.Name,
		}
	}
	return tags
}

func ConvertGetEventRow(e db.GetEventRow, isOwner, isParticipant bool) models.Event {
	lat, _ := util.NumericToFloat64(e.Latitude)
	lon, _ := util.NumericToFloat64(e.Longitude)
	base := models.Event{
		ID:             e.ID,
		Name:           e.Name,
		Description:    e.Description,
		Capacity:       e.Capacity,
		Latitude:       lat,
		Longitude:        lon,
		Address:          e.Address,
		Date:             e.Date,
		OwnerUsername:    safeString(e.OwnerUsername),
		Tags:             convertTags(e.Tags),
		IsPrivate:        e.IsPrivate,
		IsPremium:        e.IsPremium,
		CreatedAt:        e.CreatedAt,
		IsOwner:          isOwner,
		IsParticipant:    isParticipant,
		ParticipantCount: int(e.ParticipantsCount),
		ImagePath:        safeString(e.EventImagePath),
		OwnerImagePath:   safeString(e.UserImagePath),
	}

	return base
}

func ConvertListEventsRow(e db.ListEventsRow) models.Event {
	lat, _ := util.NumericToFloat64(e.Latitude)
	lon, _ := util.NumericToFloat64(e.Longitude)
	base := models.Event{
		ID:             e.ID,
		Name:           e.Name,
		Description:    e.Description,
		Capacity:       e.Capacity,
		Latitude:       lat,
		Longitude:      lon,
		Address:        e.Address,
		Date:           e.Date,
		OwnerUsername:  safeString(e.OwnerUsername),
		Tags:           convertTags(e.Tags),
		IsPrivate:      e.IsPrivate,
		IsPremium:      e.IsPremium,
		CreatedAt:      e.CreatedAt,
		IsOwner:        false,
		IsParticipant:  false,
		ParticipantCount: int(e.ParticipantsCount),
		ImagePath:      safeString(e.EventImagePath),
		OwnerImagePath: safeString(e.UserImagePath),
	}

	return base
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
) models.HomeEvents {
	return models.HomeEvents{
		Premium:     convertEventType(premium),
		Recommended: convertRecommendedEvents(recommended),
		Latest:      convertEventType(latest),
		Popular:     convertEventType(popular),
	}
}

func convertRecommendedEvents[T RecommendedRow](rows []T) []models.Event {
	result := make([]models.Event, len(rows))
	for i, row := range rows {
		result[i] = convertSingleRecommended(row)
	}
	return result
}

func convertSingleRecommended[T RecommendedRow](row T) models.Event {
	switch v := any(row).(type) {
	case db.GetUserRecommendedEventsRow:
		return convertUserRecommended(v)
	case db.GetGuestRecommendedEventsRow:
		return convertGuestRecommended(v)
	default:
		panic("unsupported type")
	}
}

func convertUserRecommended(e db.GetUserRecommendedEventsRow) models.Event {
	lat, _ := util.NumericToFloat64(e.Latitude)
	lon, _ := util.NumericToFloat64(e.Longitude)
	base := models.Event{
		ID:             e.ID,
		Name:           e.Name,
		Description:    e.Description,
		Capacity:       e.Capacity,
		Latitude:       lat,
		Longitude:      lon,
		Address:        e.Address,
		Date:           e.Date,
		OwnerUsername:  safeString(e.OwnerUsername),
		Tags:           convertTags(e.Tags),
		IsPrivate:      e.IsPrivate,
		IsPremium:      e.IsPremium,
		CreatedAt:      e.CreatedAt,
		ParticipantCount: int(e.ParticipantsCount),
		ImagePath:      safeString(e.EventImagePath),
		OwnerImagePath: safeString(e.UserImagePath),
	}

	return base
}

func convertGuestRecommended(e db.GetGuestRecommendedEventsRow) models.Event {
	lat, _ := util.NumericToFloat64(e.Latitude)
	lon, _ := util.NumericToFloat64(e.Longitude)
	base := models.Event{
		ID:             e.ID,
		Name:           e.Name,
		Description:    e.Description,
		Capacity:       e.Capacity,
		Latitude:       lat,
		Longitude:      lon,
		Address:        e.Address,
		Date:           e.Date,
		OwnerUsername:  safeString(e.OwnerUsername),
		Tags:           convertTags(e.Tags),
		IsPrivate:      e.IsPrivate,
		IsPremium:      e.IsPremium,
		CreatedAt:      e.CreatedAt,
		ParticipantCount: int(e.ParticipantsCount),
		ImagePath:      safeString(e.EventImagePath),
		OwnerImagePath: safeString(e.UserImagePath),
	}

	return base
}

func convertEventType[T any](rows []T) []models.Event {
	result := make([]models.Event, len(rows))
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
		case db.ListAllEventsRow:
			result[i] = convertListAllEventsRow(v)
		}
	}
	return result
}

func convertListAllEventsRow(e db.ListAllEventsRow) models.Event {
	lat, _ := util.NumericToFloat64(e.Latitude)
	lon, _ := util.NumericToFloat64(e.Longitude)
	base := models.Event{
		ID:             e.ID,
		Name:           e.Name,
		Description:    e.Description,
		Capacity:       e.Capacity,
		Latitude:       lat,
		Longitude:      lon,
		Address:        e.Address,
		Date:           e.Date,
		OwnerUsername:  safeString(e.OwnerUsername),
		Tags:           convertTags(e.Tags),
		IsPrivate:      e.IsPrivate,
		IsPremium:      e.IsPremium,
		CreatedAt:      e.CreatedAt,
		ParticipantCount: int(e.ParticipantsCount),
		ImagePath:      safeString(e.EventImagePath),
		OwnerImagePath: safeString(e.UserImagePath),
	}

	return base
}

func convertOwnedEventsRow(e db.GetOwnedUserEventsRow) models.Event {
	lat, _ := util.NumericToFloat64(e.Latitude)
	lon, _ := util.NumericToFloat64(e.Longitude)
	base := models.Event{
		ID:             e.ID,
		Name:           e.Name,
		Description:    e.Description,
		Capacity:       e.Capacity,
		Latitude:       lat,
		Longitude:      lon,
		Address:        e.Address,
		Date:           e.Date,
		OwnerUsername:  safeString(e.OwnerUsername),
		Tags:           convertTags(e.Tags),
		IsPrivate:      e.IsPrivate,
		IsPremium:      e.IsPremium,
		CreatedAt:      e.CreatedAt,
		ParticipantCount: int(e.ParticipantsCount),
		ImagePath:      safeString(e.EventImagePath),
		OwnerImagePath: safeString(e.UserImagePath),
	}

	return base
}

func convertPastEventsRow(e db.GetPastUserEventsRow) models.Event {
	lat, _ := util.NumericToFloat64(e.Latitude)
	lon, _ := util.NumericToFloat64(e.Longitude)
	base := models.Event{
		ID:             e.ID,
		Name:           e.Name,
		Description:    e.Description,
		Capacity:       e.Capacity,
		Latitude:       lat,
		Longitude:      lon,
		Address:        e.Address,
		Date:           e.Date,
		OwnerUsername:  safeString(e.OwnerUsername),
		Tags:           convertTags(e.Tags),
		IsPrivate:      e.IsPrivate,
		IsPremium:      e.IsPremium,
		CreatedAt:      e.CreatedAt,
		ParticipantCount: int(e.ParticipantsCount),
		ImagePath:      safeString(e.EventImagePath),
		OwnerImagePath: safeString(e.UserImagePath),
	}

	return base
}

func convertUpcomingEventsRow(e db.GetUpcomingUserEventsRow) models.Event {
	lat, _ := util.NumericToFloat64(e.Latitude)
	lon, _ := util.NumericToFloat64(e.Longitude)
	base := models.Event{
		ID:             e.ID,
		Name:           e.Name,
		Description:    e.Description,
		Capacity:       e.Capacity,
		Latitude:       lat,
		Longitude:      lon,
		Address:        e.Address,
		Date:           e.Date,
		OwnerUsername:  safeString(e.OwnerUsername),
		Tags:           convertTags(e.Tags),
		IsPrivate:      e.IsPrivate,
		IsPremium:      e.IsPremium,
		CreatedAt:      e.CreatedAt,
		ParticipantCount: int(e.ParticipantsCount),
		ImagePath:      safeString(e.EventImagePath),
		OwnerImagePath: safeString(e.UserImagePath),
	}

	return base
}

func convertPremiumEvent(e db.GetPremiumEventsRow) models.Event {
	lat, _ := util.NumericToFloat64(e.Latitude)
	lon, _ := util.NumericToFloat64(e.Longitude)
	base := models.Event{
		ID:             e.ID,
		Name:           e.Name,
		Description:    e.Description,
		Capacity:       e.Capacity,
		Latitude:       lat,
		Longitude:      lon,
		Address:        e.Address,
		Date:           e.Date,
		OwnerUsername:  safeString(e.OwnerUsername),
		Tags:           convertTags(e.Tags),
		IsPrivate:      e.IsPrivate,
		IsPremium:      e.IsPremium,
		CreatedAt:      e.CreatedAt,
		ParticipantCount: int(e.ParticipantsCount),
		ImagePath:      safeString(e.EventImagePath),
		OwnerImagePath: safeString(e.UserImagePath),
	}

	return base
}

func convertLatestEvent(e db.GetLatestEventsRow) models.Event {
	lat, _ := util.NumericToFloat64(e.Latitude)
	lon, _ := util.NumericToFloat64(e.Longitude)
	base := models.Event{
		ID:             e.ID,
		Name:           e.Name,
		Description:    e.Description,
		Capacity:       e.Capacity,
		Latitude:       lat,
		Longitude:      lon,
		Address:        e.Address,
		Date:           e.Date,
		OwnerUsername:  safeString(e.OwnerUsername),
		Tags:           convertTags(e.Tags),
		IsPrivate:      e.IsPrivate,
		IsPremium:      e.IsPremium,
		CreatedAt:      e.CreatedAt,
		ParticipantCount: int(e.ParticipantsCount),
		ImagePath:      safeString(e.EventImagePath),
		OwnerImagePath: safeString(e.UserImagePath),
	}

	return base
}

func convertPopularEvent(e db.GetPopularEventsRow) models.Event {
	lat, _ := util.NumericToFloat64(e.Latitude)
	lon, _ := util.NumericToFloat64(e.Longitude)
	base := models.Event{
		ID:             e.ID,
		Name:           e.Name,
		Description:    e.Description,
		Capacity:       e.Capacity,
		Latitude:       lat,
		Longitude:      lon,
		Address:        e.Address,
		Date:           e.Date,
		OwnerUsername:  safeString(e.OwnerUsername),
		Tags:           convertTags(e.Tags),
		IsPrivate:      e.IsPrivate,
		IsPremium:      e.IsPremium,
		CreatedAt:      e.CreatedAt,
		ParticipantCount: int(e.ParticipantsCount),
		ImagePath:      safeString(e.EventImagePath),
		OwnerImagePath: safeString(e.UserImagePath),
	}

	return base
}
