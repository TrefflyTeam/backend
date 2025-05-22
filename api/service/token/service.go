package tokenservice

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"time"
	db "treffly/db/sqlc"
	"treffly/token"
	"treffly/util"
)

type Service struct {
	store      db.Store
	tokenMaker token.Maker
	config     util.Config
	log *zap.Logger
}

func New(store db.Store, tokenMaker token.Maker, config util.Config, log *zap.Logger) *Service {
	return &Service{
		store:      store,
		tokenMaker: tokenMaker,
		config:     config,
		log:        log,
	}
}

func (s *Service) RefreshTokens(ctx context.Context, reqRefreshToken string) (accessToken string, refreshToken string, err error) {
	reqRefreshPayload, err := s.tokenMaker.VerifyToken(reqRefreshToken)
	if err != nil {
		return "", "", err
	}

	session, err := s.store.GetSession(ctx, reqRefreshPayload.ID)
	if err != nil {
		return "", "", err
	}

	if session.IsBlocked {
		err := fmt.Errorf("blocked session")
		return "", "", err
	}

	if session.UserID != reqRefreshPayload.UserID {
		err := fmt.Errorf("incorrect session user")
		return "", "", err
	}

	if session.RefreshToken != reqRefreshToken {
		err := fmt.Errorf("mismatched session token")
		return "", "", err
	}

	if time.Now().After(session.ExpiresAt) {
		err := fmt.Errorf("expired session")
		return "", "", err
	}

	accessToken, _, err = s.tokenMaker.CreateToken(
		reqRefreshPayload.UserID,
		reqRefreshPayload.IsAdmin,
		s.config.AccessTokenDuration,
	)
	if err != nil {
		return "", "", err
	}

	refreshToken, refreshPayload, err := s.tokenMaker.CreateToken(
		reqRefreshPayload.UserID,
		reqRefreshPayload.IsAdmin,
		s.config.RefreshTokenDuration,
	)
	if err != nil {
		return "", "", err
	}

	err = s.store.UpdateSession(ctx, db.UpdateSessionParams{
		OldUuid: reqRefreshPayload.ID,
		NewUuid: refreshPayload.ID,
		RefreshToken: refreshToken,
		ExpiresAt: refreshPayload.ExpiredAt,
	})
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *Service) ValidateSession(ctx context.Context, refreshToken string) error {
	payload, err := s.tokenMaker.VerifyToken(refreshToken)
	if err != nil {
		return err
	}

	session, err := s.store.GetSession(ctx, payload.ID)
	if err != nil {
		return err
	}

	if session.IsBlocked {
		err = fmt.Errorf("blocked session")
		return err
	}

	if session.UserID != payload.UserID {
		err = fmt.Errorf("incorrect session user")
		return err
	}

	if session.RefreshToken != refreshToken {
		err = fmt.Errorf("mismatched session token")
		return err
	}

	if time.Now().After(session.ExpiresAt) {
		err = fmt.Errorf("expired session")
		return err
	}

	return nil
}

func (s *Service) CreatePrivateEventToken(ctx context.Context, eventID int32, userID int32) (string, error) {
	event, err := s.store.GetEvent(ctx, db.GetEventParams{ID: eventID, OwnerID: userID})
	if err != nil {
		return "", err
	}

	if userID != event.OwnerID {
		err = fmt.Errorf("not allowed")
		return "", err
	}

	t,payload, err := s.tokenMaker.CreateToken(0, false, time.Hour)
	if err != nil {
		return "", err
	}

	arg := db.CreatePrivateEventTokenParams{
		EventID: eventID,
		Token: t,
		ExpiresAt: payload.ExpiredAt,
	}

	err = s.store.CreatePrivateEventToken(ctx, arg)
	if err != nil {
		return "", err
	}

	return t, nil
}