package eventservice

import (
	"context"
	"fmt"
	"treffly/apperror"
	db "treffly/db/sqlc"
)

type Service struct {
	store db.Store
}

func New(store db.Store) *Service {
	return &Service{store: store}
}

func (s *Service) Create(ctx context.Context, params CreateParams) (db.EventRow, error) {
	eventArg := db.CreateEventTxParams{
		Name:        params.Name,
		Description: params.Description,
		Capacity:    params.Capacity,
		Latitude:    params.Latitude,
		Longitude:   params.Longitude,
		Address:     params.Address,
		Date:        params.Date,
		IsPrivate:   params.IsPrivate,
		OwnerID:     params.OwnerID,
		Tags:        params.Tags,
		ImageID:     params.ImageID,
	}

	imageArg := db.CreateImageParams{
		ID: params.ImageID,
		Path: params.Path,
	}

	event, err := s.store.CreateEventTx(ctx, eventArg, imageArg)
	if err != nil {
		return nil, err
	}

	return event, nil
}

func (s *Service) List(ctx context.Context, params ListParams) ([]db.EventRow, error) {
	arg := db.ListEventsParams{
		UserLat:    params.Lat,
		UserLon:    params.Lon,
		SearchTerm: params.Search,
		TagIds:     params.TagIDs,
		DateRange:  params.DateRange,
	}

	rows, err := s.store.ListEvents(ctx, arg)
	if err != nil {
		return nil, err
	}

	return db.ConvertToEventRow(rows), nil
}

func (s *Service) Update(ctx context.Context, params UpdateParams) (db.EventRow, error) {
	event, err := s.store.GetEvent(ctx, params.EventID)
	if err != nil {
		return nil, err
	}

	if event.OwnerID != params.UserID {
		err = fmt.Errorf("owner id missmatch")
		return nil, apperror.Forbidden.WithCause(err)
	}

	arg := db.UpdateEventTxParams{
		EventID:     params.EventID,
		Name:        params.Name,
		Description: params.Description,
		Capacity:    params.Capacity,
		Latitude:    params.Latitude,
		Longitude:   params.Longitude,
		Address:     params.Address,
		Date:        params.Date,
		IsPrivate:   params.IsPrivate,
		Tags:        params.Tags,
	}

	updatedEvent, err := s.store.UpdateEventTx(ctx, arg)
	if err != nil {
		return nil, err
	}

	return updatedEvent, nil
}

func (s *Service) Delete(ctx context.Context, params DeleteParams) error {
	event, err := s.store.GetEvent(ctx, params.EventID)
	if err != nil {
		return err
	}

	if event.OwnerID != params.UserID {
		err = fmt.Errorf("owner id missmatch")
		return apperror.Forbidden.WithCause(err)
	}

	return s.store.DeleteEvent(ctx, params.EventID)
}

func (s *Service) GetHome(ctx context.Context, params GetHomeParams) (*HomeEvents, error) {
	result := &HomeEvents{}
	var err error

	result.Premium, err = s.getPremiumEvents(ctx)
	if err != nil {
		return nil, err
	}

	result.Recommended, err = s.getRecommendedEvents(ctx, params)
	if err != nil {
		return nil, err
	}

	result.Latest, err = s.getLatestEvents(ctx)
	if err != nil {
		return nil, err
	}

	result.Popular, err = s.getPopularEvents(ctx)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *Service) getPremiumEvents(ctx context.Context) ([]db.EventRow, error) {
	rows, err := s.store.GetPremiumEvents(ctx)
	if err != nil {
		return nil, err
	}
	return db.ConvertToEventRow(rows), nil
}

func (s *Service) getRecommendedEvents(ctx context.Context, params GetHomeParams) ([]db.EventRow, error) {
	if params.UserID > 0 {
		arg := db.GetUserRecommendedEventsParams{
			UserID:  params.UserID,
			UserLat: params.Lat,
			UserLon: params.Lon,
		}

		rows, err := s.store.GetUserRecommendedEvents(ctx, arg)
		if err != nil {
			return nil, err
		}
		return db.ConvertToEventRow(rows), nil
	}

	arg := db.GetGuestRecommendedEventsParams{
		UserLat: params.Lat,
		UserLon: params.Lon,
	}

	rows, err := s.store.GetGuestRecommendedEvents(ctx, arg)
	if err != nil {
		return nil, err
	}
	return db.ConvertToEventRow(rows), nil
}

func (s *Service) getLatestEvents(ctx context.Context) ([]db.EventRow, error) {
	rows, err := s.store.GetLatestEvents(ctx)
	if err != nil {
		return nil, err
	}
	return db.ConvertToEventRow(rows), nil
}

func (s *Service) getPopularEvents(ctx context.Context) ([]db.EventRow, error) {
	rows, err := s.store.GetPopularEvents(ctx)
	if err != nil {
		return nil, err
	}
	return db.ConvertToEventRow(rows), nil
}

func (s *Service) Subscribe(ctx context.Context, params SubscriptionParams) (EventWithStatus, error) {
	arg := db.SubscribeToEventParams{
		EventID: params.EventID,
		UserID:  params.UserID,
	}

	if err := s.store.SubscribeToEvent(ctx, arg); err != nil {
		return EventWithStatus{}, err
	}

	return s.GetEventWithStatus(ctx, params.EventID, params.UserID)
}

func (s *Service) Unsubscribe(ctx context.Context, params SubscriptionParams) (EventWithStatus, error) {
	arg := db.UnsubscribeFromEventParams{
		EventID: params.EventID,
		UserID:  params.UserID,
	}

	if err := s.store.UnsubscribeFromEvent(ctx, arg); err != nil {
		return EventWithStatus{}, err
	}

	return s.GetEventWithStatus(ctx, params.EventID, params.UserID)
}

func (s *Service) GetEventWithStatus(ctx context.Context, eventID, userID int32) (EventWithStatus, error) {
	event, err := s.store.GetEvent(ctx, eventID)
	if err != nil {
		return EventWithStatus{}, err
	}

	participantArg := db.IsParticipantParams{
		EventID: eventID,
		UserID:  userID,
	}

	isParticipant, err := s.store.IsParticipant(ctx, participantArg)
	if err != nil {
		return EventWithStatus{}, err
	}

	return EventWithStatus{
		Event:         event,
		IsOwner:       event.OwnerID == userID,
		IsParticipant: isParticipant,
	}, nil
}

func (s *Service) GetUpcomingUserEvents(ctx context.Context, userID int32) ([]db.EventRow, error) {
	rows, err := s.store.GetUpcomingUserEvents(ctx, userID)
	if err != nil {
		return nil, err
	}
	return db.ConvertToEventRow(rows), nil
}

func (s *Service) GetPastUserEvents(ctx context.Context, userID int32) ([]db.EventRow, error) {
	rows, err := s.store.GetPastUserEvents(ctx, userID)
	if err != nil {
		return nil, err
	}
	return db.ConvertToEventRow(rows), nil
}

func (s *Service) GetOwnedUserEvents(ctx context.Context, userID int32) ([]db.EventRow, error) {
	rows, err := s.store.GetOwnedUserEvents(ctx, userID)
	if err != nil {
		return nil, err
	}
	return db.ConvertToEventRow(rows), nil
}
