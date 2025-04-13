package db

import (
	"context"
	"fmt"
)

type UpdateUserTagsTxParams struct {
	UserID int32   `json:"user_id"`
	Tags   []int32 `json:"tags"`
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
