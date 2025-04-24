package db

import (
	"context"
	"fmt"
	"github.com/google/uuid"
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
	ImageID     uuid.UUID      `json:"image_id"`
}

func (store *SQLStore) CreateEventTx(ctx context.Context, eventParams CreateEventTxParams, imageParams CreateImageParams) (GetEventRow, error) {
	var result GetEventRow

	err := store.execTx(ctx, func(q *Queries) error {
		_, err := q.CreateImage(ctx, imageParams)
		if err != nil {
			return fmt.Errorf("create image error: %w", err)
		}

		event, err := q.CreateEvent(ctx, CreateEventParams{
			Name:        eventParams.Name,
			Description: eventParams.Description,
			Capacity:    eventParams.Capacity,
			Latitude:    eventParams.Latitude,
			Longitude:   eventParams.Longitude,
			Address:     eventParams.Address,
			Date:        eventParams.Date,
			OwnerID:     eventParams.OwnerID,
			IsPrivate:   eventParams.IsPrivate,
			ImageID:     eventParams.ImageID,
		})
		if err != nil {
			return fmt.Errorf("create event error: %w", err)
		}

		for _, tagID := range eventParams.Tags {
			if _, err = q.AddEventTag(ctx, AddEventTagParams{
				EventID: event.ID,
				TagID:   tagID,
			}); err != nil {
				return fmt.Errorf("add tag %d error: %w", tagID, err)
			}
		}

		result, err = q.GetEvent(ctx, event.ID)
		if err != nil {
			return fmt.Errorf("get event with tags error: %w", err)
		}

		return nil
	})

	if err != nil {
		return GetEventRow{}, fmt.Errorf("transaction failed: %w", err)
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
	ImageID     uuid.UUID
}

func (store *SQLStore) UpdateEventTx(ctx context.Context, arg UpdateEventTxParams) (GetEventRow, error) {
	var result GetEventRow

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
			ImageID:     arg.ImageID,
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

		result, err = q.GetEvent(ctx, arg.EventID)
		if err != nil {
			return fmt.Errorf("get updated event error: %w", err)
		}

		return nil
	})

	return result, err
}
