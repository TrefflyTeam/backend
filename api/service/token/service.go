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
		s.log.Error("Error getting session",
			zap.Any("session id", reqRefreshPayload.ID),
			zap.String("token prefix", reqRefreshToken[:20]),
			zap.Error(err))
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
		s.log.Error("Error updating session",
			zap.Any("session id", reqRefreshPayload.ID),
			zap.String("token prefix", reqRefreshToken[:20]),
			zap.Error(err))
		return "", "", err
	}

	return accessToken, refreshToken, nil
}


