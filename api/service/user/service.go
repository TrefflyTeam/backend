package userservice

import (
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
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

func (s *Service) CreateUser(ctx context.Context, params CreateParams) (*User, error) {
	hashedPassword, err := util.HashPassword(params.Password)
	if err != nil {
		return nil, apperror.InternalServer.WithCause(err)
	}

	user, err := s.store.CreateUser(ctx, db.CreateUserParams{
		Username:     params.Username,
		Email:        params.Email,
		PasswordHash: hashedPassword,
	})

	resp := ConvertUser(user)

	return &resp, err
}

func (s *Service) LoginUser(ctx context.Context, email, password string) (*User, string, string, error) {
	user, err := s.store.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", "", apperror.InvalidCredentials.WithCause(err)
		}
		return nil, "", "", err
	}

	if err := util.CheckPassword(password, user.PasswordHash); err != nil {
		return nil, "", "", apperror.InvalidCredentials.WithCause(err)
	}

	accessToken, _, err := s.tokenMaker.CreateToken(user.ID, s.config.AccessTokenDuration)
	if err != nil {
		return nil, "", "", apperror.InternalServer.WithCause(err)
	}

	refreshToken, refreshPayload, err := s.tokenMaker.CreateToken(user.ID, s.config.RefreshTokenDuration)
	if err != nil {
		return nil, "", "", apperror.InternalServer.WithCause(err)
	}

	err = s.store.CreateSession(ctx, db.CreateSessionParams{
		Uuid:         refreshPayload.ID,
		UserID:       user.ID,
		RefreshToken: refreshToken,
		ExpiresAt:    refreshPayload.ExpiredAt,
		IsBlocked:    false,
	})
	if err != nil {
		return nil, "", "", err
	}

	resp := ConvertUser(user)

	return &resp, accessToken, refreshToken, nil
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

func (s *Service) GetUserWithTags(ctx context.Context, userID int32) (*UserWithTags, error) {
	user, err := s.store.GetUserWithTags(ctx, userID)

	resp := ConvertUserWithTags(user)

	return &resp, err
}

func (s *Service) UpdateUser(ctx context.Context, params UpdateUserParams) (*UserWithTags, error) {
	user, err := s.store.GetUser(ctx, params.ID)
	if err != nil {
		return nil, err
	}

	imageID := params.NewImageID
	path := params.Path
	if params.DeleteImage {
		imageID = uuid.Nil
		path = ""
	}
	if !params.DeleteImage && params.NewImageID == uuid.Nil {
		imageID = user.ImageID.Bytes
	}

	arg := db.UpdateUserTxParams{
		UserID:       params.ID,
		Username:     params.Username,
		NewImageID:   imageID,
		NewPath:      path,
		OldImageID:   user.ImageID.Bytes,
	}

	updatedUser, err := s.store.UpdateUserTx(ctx, arg)
	if err != nil {
		return nil, err
	}

	resp := ConvertUserWithTags(updatedUser)

	return &resp, err
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
