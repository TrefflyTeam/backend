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
	Tags        []int32        `json:"tags"`
}

func (store *SQLStore) CreateEventTx(ctx context.Context, params CreateEventTxParams) (EventResponse, error) {
	var result EventResponse

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

		result = ConvertRowToEvent(fullEvent)
		return nil
	})

	if err != nil {
		return EventResponse{}, fmt.Errorf("transaction failed: %w", err)
	}

	return result, nil
}

type UpdateEventTxParams struct {
	EventID     int32
	Name        string
	Description string
	Capacity    int32
	Latitude    pgtype.Numeric
	Longitude   pgtype.Numeric
	Address     string
	Date        time.Time
	IsPrivate   bool
	Tags        []int32
}


func (store *SQLStore) UpdateEventTx(ctx context.Context, arg UpdateEventTxParams) (EventResponse, error) {
	var result EventResponse

	err := store.execTx(ctx, func(q *Queries) error {
		err := q.UpdateEvent(ctx, UpdateEventParams{
			ID:          arg.EventID,
			Name:        arg.Name,
			Description: arg.Description,
			Capacity:    arg.Capacity,
			Latitude:    arg.Latitude,
			Longitude:   arg.Longitude,
			Address:     arg.Address,
			Date:        arg.Date,
			IsPrivate:   arg.IsPrivate,
		})
		if err != nil {
			return fmt.Errorf("update event error: %w", err)
		}

		err = q.DeleteAllEventTags(ctx, arg.EventID)
		if err != nil {
			return fmt.Errorf("delete old tags error: %w", err)
		}

		for _, tagID := range arg.Tags {
			_, err := q.AddEventTag(ctx, AddEventTagParams{
				EventID: arg.EventID,
				TagID:   tagID,
			})
			if err != nil {
				return fmt.Errorf("add new tag %d error: %w", tagID, err)
			}
		}

		fullEvent, err := q.GetEvent(ctx, arg.EventID)
		if err != nil {
			return fmt.Errorf("get updated event error: %w", err)
		}

		result = ConvertRowToEvent(fullEvent)
		return nil
	})

	return result, err
}
