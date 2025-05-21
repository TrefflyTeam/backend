package userservice

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"math"
	"strconv"
	"treffly/api/models"
	"treffly/apperror"
	"treffly/db/redis"
	"treffly/db/sqlc"
	"treffly/token"
	"treffly/util"
)

type Service struct {
	store          db.Store
	resetStore     redis.ResetStore
	rateLimitStore redis.RateLimitStore
	tokenMaker     token.Maker
	config         util.Config
}

func New(store db.Store, redisStore redis.ResetStore, tokenMaker token.Maker, config util.Config, rateLimitStore redis.RateLimitStore) *Service {
	return &Service{
		store:      store,
		tokenMaker: tokenMaker,
		config:     config,
		resetStore: redisStore,
		rateLimitStore: rateLimitStore,
	}
}

func (s *Service) CreateUser(ctx context.Context, params models.CreateUserParams) (models.User, error) {
	hashedPassword, err := util.HashPassword(params.Password)
	if err != nil {
		return models.User{}, apperror.InternalServer.WithCause(err)
	}

	user, err := s.store.CreateUser(ctx, db.CreateUserParams{
		Username:     params.Username,
		Email:        params.Email,
		PasswordHash: hashedPassword,
	})

	resp := ConvertUser(user)

	return resp, err
}

func (s *Service) LoginUser(ctx context.Context, email, password string) (models.User, string, string, error) {
	user, err := s.store.GetUserByEmail(ctx, email)
	if err != nil {
		return models.User{}, "", "", err
	}

	if err := util.CheckPassword(password, user.PasswordHash); err != nil {
		return models.User{}, "", "", apperror.InvalidCredentials.WithCause(err)
	}

	accessToken, _, err := s.tokenMaker.CreateToken(user.ID, false, s.config.AccessTokenDuration)
	if err != nil {
		return models.User{}, "", "", apperror.InternalServer.WithCause(err)
	}

	refreshToken, refreshPayload, err := s.tokenMaker.CreateToken(user.ID, false,  s.config.RefreshTokenDuration)
	if err != nil {
		return models.User{}, "", "", apperror.InternalServer.WithCause(err)
	}

	err = s.store.CreateSession(ctx, db.CreateSessionParams{
		Uuid:         refreshPayload.ID,
		UserID:       user.ID,
		RefreshToken: refreshToken,
		ExpiresAt:    refreshPayload.ExpiredAt,
		IsBlocked:    false,
	})
	if err != nil {
		return models.User{}, "", "", err
	}

	resp := ConvertUser(user)

	return resp, accessToken, refreshToken, nil
}

func (s *Service) CreateAuthSession(ctx context.Context, userID int32) (string, string, error) {
	accessToken, _, err := s.tokenMaker.CreateToken(userID, false,  s.config.AccessTokenDuration)
	if err != nil {
		return "", "", apperror.InternalServer.WithCause(err)
	}

	refreshToken, refreshPayload, err := s.tokenMaker.CreateToken(userID, false, s.config.RefreshTokenDuration)
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

func (s *Service) GetUserWithTags(ctx context.Context, userID int32) (models.UserWithTags, error) {
	user, err := s.store.GetUserWithTags(ctx, userID)

	resp := ConvertUserWithTags(user)

	return resp, err
}

func (s *Service) UpdateUser(ctx context.Context, params models.UpdateUserParams) (models.UserWithTags, error) {
	user, err := s.store.GetUser(ctx, params.ID)
	if err != nil {
		return models.UserWithTags{}, err
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
		UserID:     params.ID,
		Username:   params.Username,
		NewImageID: imageID,
		NewPath:    path,
		OldImageID: user.ImageID.Bytes,
	}

	updatedUser, err := s.store.UpdateUserTx(ctx, arg)
	if err != nil {
		return models.UserWithTags{}, err
	}

	resp := ConvertUserWithTags(updatedUser)

	return resp, err
}

func (s *Service) UpdateUserTags(ctx context.Context, params models.UpdateUserTagsParams) error {
	return s.store.UpdateUserTagsTx(ctx, db.UpdateUserTagsTxParams{
		UserID: params.UserID,
		Tags:   params.TagIDs,
	})
}

func (s *Service) DeleteUser(ctx context.Context, userID int32) error {
	return s.store.DeleteUser(ctx, userID)
}

func (s *Service) InitiatePasswordReset(ctx context.Context, email string) (string, error) {
	user, err := s.store.GetUserByEmail(ctx, email)
	if err != nil {
		return "", nil
	}

	allowed, err := s.rateLimitStore.CanSendResetRequest(ctx, email, s.config.SendCodeRateLimit)
	if err != nil {
		return "", err
	}
	if !allowed {
		return "", errors.New("too many requests")
	}

	code, err := s.generateResetCode()
	if err != nil {
		return "", err
	}

	err = s.resetStore.SaveCode(ctx, string(user.ID), code, s.config.ResetCodeTTL)
	if err != nil {
		return "", err
	}

	return code, nil
}

func (s *Service) ConfirmResetCode(ctx context.Context, email, code string) (string, error) {
	user, err := s.store.GetUserByEmail(ctx, email)
	if err != nil {
		return "", nil
	}

	storedCode, err := s.resetStore.GetCode(ctx, string(user.ID))
	if err != nil {
		return "", err
	}

	if storedCode != code {
		return "", errors.New("invalid code")
	}

	t, _, err := s.tokenMaker.CreateToken(user.ID, false, s.config.ResetTokenDuration)
	if err != nil {
		return "", err
	}

	id := strconv.Itoa(int(user.ID))

	if err := s.resetStore.SaveToken(ctx, t, id, s.config.ResetTokenDuration); err != nil {
		return "", err
	}

	_ = s.resetStore.DeleteCode(ctx, id) //TODO: log error

	return t, nil
}

func (s *Service) ValidateResetToken(ctx context.Context, token string) (string, error) {
	_, err := s.tokenMaker.VerifyToken(token)
	if err != nil {
		return "", err
	}

	userID, err := s.resetStore.GetToken(ctx, token)
	if err != nil {
		return "", err
	}

	return userID, nil
}

func (s *Service) CompletePasswordReset(ctx context.Context, token, newPassword string) error {
	userID, err := s.ValidateResetToken(ctx, token)
	if err != nil {
		return err
	}
	id, err := strconv.Atoi(userID)
	if err != nil {
		return err
	}

	hashedPassword, err := util.HashPassword(newPassword)
	if err != nil {
		return err
	}

	err = s.store.UpdatePassword(ctx, db.UpdatePasswordParams{
		ID:           int32(id),
		PasswordHash: hashedPassword,
	})
	if err != nil {
		return err
	}

	_ = s.resetStore.DeleteToken(ctx, token) //TODO: log
	return nil
}

func (s *Service) generateResetCode() (string, error) {
	if s.config.ResetCodeLength < 1 || s.config.ResetCodeLength > 9 {
		return "", fmt.Errorf("invalid code length: %d", s.config.ResetCodeLength)
	}

	var buf [4]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", fmt.Errorf("failed to generate code: %w", err)
	}

	num := binary.BigEndian.Uint32(buf[:])
	maxValue := uint32(math.Pow10(s.config.ResetCodeLength))
	code := num % maxValue

	format := fmt.Sprintf("%%0%dd", s.config.ResetCodeLength)
	return fmt.Sprintf(format, code), nil
}
