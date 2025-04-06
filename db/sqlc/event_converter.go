package db

import (
	"github.com/jackc/pgx/v5/pgtype"
	"time"
)

type EventResponse struct {
	ID                int32          `json:"id"`
	Name              string         `json:"name"`
	Description       string         `json:"description"`
	Capacity          int32          `json:"capacity"`
	Latitude          pgtype.Numeric `json:"latitude"`
	Longitude         pgtype.Numeric `json:"longitude"`
	Address           string         `json:"address"`
	Date              time.Time      `json:"date"`
	OwnerUsername     string    `json:"owner_username"`
	IsPrivate         bool           `json:"is_private"`
	IsPremium         bool           `json:"is_premium"`
	CreatedAt         time.Time      `json:"created_at"`
	Tags              []Tag          `json:"tags"`
	ParticipantsCount int64          `json:"participants_count"`
}

func ConvertRowToEvent(row interface{}) EventResponse {
	switch v := row.(type) {
	case GetUserRecommendedEventsRow:
		return EventResponse{
			ID:                v.ID,
			Name:              v.Name,
			Description:       v.Description,
			Capacity:          v.Capacity,
			Latitude:          v.Latitude,
			Longitude:         v.Longitude,
			Address:           v.Address,
			Date:              v.Date,
			OwnerUsername:     v.OwnerUsername.String,
			IsPrivate:         v.IsPrivate,
			IsPremium:         v.IsPremium,
			CreatedAt:         v.CreatedAt,
			Tags:              v.Tags,
			ParticipantsCount: v.ParticipantsCount,
		}
	case GetGuestRecommendedEventsRow:
		return EventResponse{
			ID:                v.ID,
			Name:              v.Name,
			Description:       v.Description,
			Capacity:          v.Capacity,
			Latitude:          v.Latitude,
			Longitude:         v.Longitude,
			Address:           v.Address,
			Date:              v.Date,
			OwnerUsername:     v.OwnerUsername.String,
			IsPrivate:         v.IsPrivate,
			IsPremium:         v.IsPremium,
			CreatedAt:         v.CreatedAt,
			Tags:              v.Tags,
			ParticipantsCount: v.ParticipantsCount,
		}
	case GetPopularEventsRow:
		return EventResponse{
			ID:                v.ID,
			Name:              v.Name,
			Description:       v.Description,
			Capacity:          v.Capacity,
			Latitude:          v.Latitude,
			Longitude:         v.Longitude,
			Address:           v.Address,
			Date:              v.Date,
			OwnerUsername:     v.OwnerUsername.String,
			IsPrivate:         v.IsPrivate,
			IsPremium:         v.IsPremium,
			CreatedAt:         v.CreatedAt,
			Tags:              v.Tags,
			ParticipantsCount: v.ParticipantsCount,
		}
	case GetLatestEventsRow:
		return EventResponse{
			ID:                v.ID,
			Name:              v.Name,
			Description:       v.Description,
			Capacity:          v.Capacity,
			Latitude:          v.Latitude,
			Longitude:         v.Longitude,
			Address:           v.Address,
			Date:              v.Date,
			OwnerUsername:     v.OwnerUsername.String,
			IsPrivate:         v.IsPrivate,
			IsPremium:         v.IsPremium,
			CreatedAt:         v.CreatedAt,
			Tags:              v.Tags,
			ParticipantsCount: v.ParticipantsCount,
		}
	case GetPremiumEventsRow:
		return EventResponse{
			ID:                v.ID,
			Name:              v.Name,
			Description:       v.Description,
			Capacity:          v.Capacity,
			Latitude:          v.Latitude,
			Longitude:         v.Longitude,
			Address:           v.Address,
			Date:              v.Date,
			OwnerUsername:     v.OwnerUsername.String,
			IsPrivate:         v.IsPrivate,
			IsPremium:         v.IsPremium,
			CreatedAt:         v.CreatedAt,
			Tags:              v.Tags,
			ParticipantsCount: v.ParticipantsCount,
		}
	case GetEventRow:
		return EventResponse{
			ID:                v.ID,
			Name:              v.Name,
			Description:       v.Description,
			Capacity:          v.Capacity,
			Latitude:          v.Latitude,
			Longitude:         v.Longitude,
			Address:           v.Address,
			Date:              v.Date,
			OwnerUsername:     v.OwnerUsername.String,
			IsPrivate:         v.IsPrivate,
			IsPremium:         v.IsPremium,
			CreatedAt:         v.CreatedAt,
			Tags:              v.Tags,
			ParticipantsCount: v.ParticipantsCount,
		}
	case ListEventsRow:
		return EventResponse{
			ID:                v.ID,
			Name:              v.Name,
			Description:       v.Description,
			Capacity:          v.Capacity,
			Latitude:          v.Latitude,
			Longitude:         v.Longitude,
			Address:           v.Address,
			Date:              v.Date,
			OwnerUsername:     v.OwnerUsername.String,
			IsPrivate:         v.IsPrivate,
			IsPremium:         v.IsPremium,
			CreatedAt:         v.CreatedAt,
			Tags:              v.Tags,
			ParticipantsCount: v.ParticipantsCount,
		}
	default:
		panic("unsupported type")
	}
}
