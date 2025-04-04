package db

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store interface {
	Querier
	CreateEventTx(ctx context.Context, params CreateEventTxParams) (CreateEventTxResult, error)
}

type SQLStore struct {
	db *pgxpool.Pool
	*Queries
}

func NewStore(db *pgxpool.Pool) Store {
	return &SQLStore{db: db, Queries: New(db)}
}
