package userservice

import (
	"context"
	"database/sql"
	"errors"
	"treffly/apperror"
	"treffly/db/sqlc"
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

func (s *Service) CreateUser(ctx context.Context, params CreateParams) (db.User, error) {
	hashedPassword, err := util.HashPassword(params.Password)
	if err != nil {
		return db.User{}, apperror.InternalServer.WithCause(err)
	}

	user, err := s.store.CreateUser(ctx, db.CreateUserParams{
		Username:     params.Username,
		Email:        params.Email,
		PasswordHash: hashedPassword,
	})

	return user, err
}

func (s *Service) LoginUser(ctx context.Context, email, password string) (db.User, string, string, error) {
	user, err := s.store.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.User{}, "", "", apperror.InvalidCredentials.WithCause(err)
		}
		return db.User{}, "", "", err
	}

	if err := util.CheckPassword(password, user.PasswordHash); err != nil {
		return db.User{}, "", "", apperror.InvalidCredentials.WithCause(err)
	}

	accessToken, _, err := s.tokenMaker.CreateToken(user.ID, s.config.AccessTokenDuration)
	if err != nil {
		return db.User{}, "", "", apperror.InternalServer.WithCause(err)
	}

	refreshToken, refreshPayload, err := s.tokenMaker.CreateToken(user.ID, s.config.RefreshTokenDuration)
	if err != nil {
		return db.User{}, "", "", apperror.InternalServer.WithCause(err)
	}

	err = s.store.CreateSession(ctx, db.CreateSessionParams{
		Uuid:         refreshPayload.ID,
		UserID:       user.ID,
		RefreshToken: refreshToken,
		ExpiresAt:    refreshPayload.ExpiredAt,
		IsBlocked:    false,
	})
	if err != nil {
		return db.User{}, "", "", err
	}

	return user, accessToken, refreshToken, nil
}

func (s *Service) CreateAuthSession(ctx context.Context, userID int32) (string, string, error) {
	accessToken, _, err := s.tokenMaker.CreateToken(userID, s.config.AccessTokenDuration)
	if err != nil {
		return "", "", apperror.InternalServer.WithCause(err)
	}

	refreshToken, refreshPayload, err := s.tokenMaker.CreateToken(userID, s.config.RefreshTokenDuration)
	if err != nil {
		return "", "", apperror.InternalServer.WithCause(err)
	}

	err = s.store.CreateSession(ctx, db.CreateSessionParams{
		Uuid:         refreshPayload.ID,
		UserID:       userID,
		RefreshToken: refreshToken,
		ExpiresAt:    refreshPayload.ExpiredAt,
		IsBlocked:    false,
	})

	return accessToken, refreshToken, err
}

func (s *Service) GetUserWithTags(ctx context.Context, userID int32) (db.UserWithTagsView, error) {
	user, err := s.store.GetUserWithTags(ctx, userID)
	return user, err
}

func (s *Service) UpdateUser(ctx context.Context, params UpdateUserParams) (db.UserWithTagsView, error) {
	_, err := s.store.UpdateUser(ctx, db.UpdateUserParams{
		ID:       params.ID,
		Username: params.Username,
	})
	if err != nil {
		return db.UserWithTagsView{}, err
	}

	user, err := s.store.GetUserWithTags(ctx, params.ID)
	if err != nil {
		return db.UserWithTagsView{}, err
	}
	return user, nil
}

func (s *Service) UpdateUserTags(ctx context.Context, params UpdateUserTagsParams) error {
	return s.store.UpdateUserTagsTx(ctx, db.UpdateUserTagsTxParams{
		UserID: params.UserID,
		Tags:   params.TagIDs,
	})
}

func (s *Service) DeleteUser(ctx context.Context, userID int32) error {
	return s.store.DeleteUser(ctx, userID)
}
