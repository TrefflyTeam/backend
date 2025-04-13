package db

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store interface {
	Querier
	CreateEventTx(ctx context.Context, params CreateEventTxParams) (GetEventRow, error)
	UpdateEventTx(ctx context.Context, params UpdateEventTxParams) (GetEventRow, error)
	UpdateUserTagsTx(ctx context.Context, params UpdateUserTagsTxParams) error
}

type SQLStore struct {
	db *pgxpool.Pool
	*Queries
}

func NewStore(db *pgxpool.Pool) Store {
	return &SQLStore{db: db, Queries: New(db)}
}
