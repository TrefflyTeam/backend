package tagservice

import (
	"context"
	"treffly/api/models"
	db "treffly/db/sqlc"
)

type Service struct {
	store      db.Store
}

func New(store db.Store) *Service {
	return &Service{
		store:      store,
	}
}

func (s *Service) GetTags(ctx context.Context) ([]models.Tag, error) {
	tags, err := s.store.GetTags(ctx)
	if err != nil {
		return nil, err
	}

	return convertTags(tags), nil
}
