package tokenservice

import (
	"context"
	"fmt"
	"time"
	db "treffly/db/sqlc"
	"treffly/token"
	"treffly/util"
)

type Service struct {
	store      db.Store
	tokenMaker token.Maker
	config     util.Config
}

func New(store db.Store, tokenMaker token.Maker, config util.Config) *Service {
	return &Service{
		store:      store,
		tokenMaker: tokenMaker,
		config:     config,
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
		s.config.AccessTokenDuration,
	)
	if err != nil {
		return "", "", err
	}

	refreshToken, refreshPayload, err := s.tokenMaker.CreateToken(
		reqRefreshPayload.UserID,
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


