package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgtype"
	"time"
)

type CreateEventTxParams struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Capacity    int32          `json:"capacity"`
	Latitude    pgtype.Numeric `json:"latitude"`
	Longitude   pgtype.Numeric `json:"longitude"`
	Address     string         `json:"address"`
	Date        time.Time      `json:"date"`
	OwnerID     int32          `json:"owner_id"`
	IsPrivate   bool           `json:"is_private"`
	Tags        []int32          `json:"tags"`
}

type CreateEventTxResult struct {
	ID          int32          `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Capacity    int32          `json:"capacity"`
	Latitude    pgtype.Numeric `json:"latitude"`
	Longitude   pgtype.Numeric `json:"longitude"`
	Address     string         `json:"address"`
	Date        time.Time      `json:"date"`
	CreatedAt   time.Time      `json:"created_at"`
	IsPrivate   bool           `json:"is_private"`
	IsPremium   bool           `json:"is_premium"`
	Tags        []Tag          `json:"tags"`
}

func (store *SQLStore) CreateEventTx(ctx context.Context, params CreateEventTxParams) (CreateEventTxResult, error) {
	var result CreateEventTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		event, err := q.CreateEvent(ctx, CreateEventParams{
			Name:        params.Name,
			Description: params.Description,
			Capacity:    params.Capacity,
			Latitude:    params.Latitude,
			Longitude:   params.Longitude,
			Address:     params.Address,
			Date:        params.Date,
			OwnerID:     params.OwnerID,
			IsPrivate:   params.IsPrivate,
		})
		if err != nil {
			return fmt.Errorf("create event error: %w", err)
		}

		for _, tagID := range params.Tags {
			if _, err = q.AddEventTag(ctx, AddEventTagParams{
				EventID: event.ID,
				TagID:   tagID,
			}); err != nil {
				return fmt.Errorf("add tag %d error: %w", tagID, err)
			}
		}

		fullEvent, err := q.GetEvent(ctx, event.ID)
		if err != nil {
			return fmt.Errorf("get event with tags error: %w", err)
		}

		result = convertToEventTxResult(fullEvent)
		return nil
	})

	if err != nil {
		return CreateEventTxResult{}, fmt.Errorf("transaction failed: %w", err)
	}

	return result, nil
}

func convertToEventTxResult(event EventWithTagsView) CreateEventTxResult {
	return CreateEventTxResult{
		ID:          event.ID,
		Name:        event.Name,
		Description: event.Description,
		Capacity:    event.Capacity,
		Latitude:    event.Latitude,
		Longitude:   event.Longitude,
		Address:     event.Address,
		Date:        event.Date,
		CreatedAt:   event.CreatedAt,
		IsPrivate:   event.IsPrivate,
		IsPremium:   event.IsPremium,
		Tags:        event.Tags,
	}
}

