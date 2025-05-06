package db

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type UpdateUserTagsTxParams struct {
	UserID int32
	Tags   []int32
}

func (store *SQLStore) UpdateUserTagsTx(ctx context.Context, params UpdateUserTagsTxParams) error {
	err := store.execTx(ctx, func(q *Queries) error {
		err := q.DeleteUserTags(ctx, params.UserID)
		if err != nil {
			return fmt.Errorf("delete user tags error: %w", err)
		}

		arg := AddUserTagsParams{
			UserID:  params.UserID,
			Tags:    params.Tags,
		}

		err = q.AddUserTags(ctx, arg)
		if err != nil {
			return fmt.Errorf("add user tags error: %w", err)
		}

		return nil
	})

	return err
}

type UpdateUserTxParams struct {
	UserID int32
	Username string
	NewImageID  uuid.UUID
	NewPath     string
	OldImageID  uuid.UUID
}

func (store *SQLStore) UpdateUserTx(ctx context.Context, params UpdateUserTxParams) (UserWithTagsView, error) {
	var result UserWithTagsView

	err := store.execTx(ctx, func(q *Queries) error {
		newImageID := params.NewImageID
		if params.NewImageID == params.OldImageID {
			newImageID = params.OldImageID
		}
		newImageUUID := pgtype.UUID{
			Bytes: newImageID,
			Valid: newImageID != uuid.Nil,
		}

		if newImageUUID.Valid && params.NewPath != "" && params.OldImageID != params.NewImageID{
			imageArg := CreateImageParams{
				params.NewImageID,
				params.NewPath,
			}
			_, err := q.CreateImage(ctx, imageArg)
			if err != nil {
				return fmt.Errorf("create image error: %w", err)
			}
		}

		_, err := q.UpdateUser(ctx, UpdateUserParams{
			ID: params.UserID,
			Username: params.Username,
			ImageID: newImageUUID,
		})
		if err != nil {
			return fmt.Errorf("update user tags error: %w", err)
		}

		oldImageUUID := pgtype.UUID{
			Bytes: params.OldImageID,
			Valid: params.OldImageID != uuid.Nil,
		}

		if oldImageUUID.Valid && params.OldImageID != params.NewImageID {
			err = q.DeleteImage(ctx, oldImageUUID.Bytes)
			if err != nil {
				return fmt.Errorf("delete old image error: %w", err)
			}
		}

		result, err = q.GetUserWithTags(ctx, params.UserID)
		if err != nil {
			return fmt.Errorf("get user tags error: %w", err)
		}

		return nil
	})

	return result, err
}
