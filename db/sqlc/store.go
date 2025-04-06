package db

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store interface {
	Querier
	CreateEventTx(ctx context.Context, params CreateEventTxParams) (EventResponse, error)
	UpdateEventTx(ctx context.Context, params UpdateEventTxParams) (EventResponse, error)
}

type SQLStore struct {
	db *pgxpool.Pool
	*Queries
}

func NewStore(db *pgxpool.Pool) Store {
	return &SQLStore{db: db, Queries: New(db)}
}
