package redis

import (
	"context"
	"fmt"
	"time"
)

type ResetStore struct {
	client *Client
}

func NewRedisResetStore(client *Client) ResetStore {
	return ResetStore{
		client: client,
	}
}

func (s *ResetStore) codeKey(userID string) string {
	return fmt.Sprintf("reset:code:%s", userID)
}

func (s *ResetStore) tokenKey(token string) string {
	return fmt.Sprintf("reset:token:%s", token)
}

func (s *ResetStore) SaveCode(ctx context.Context, userID, code string, resetCodeTTL time.Duration) error {
	return s.client.SetEx(
		ctx,
		s.codeKey(userID),
		code,
		resetCodeTTL,
	).Err()
}

func (s *ResetStore) GetCode(ctx context.Context, userID string) (string, error) {
	return s.client.Get(ctx, s.codeKey(userID)).Result()
}

func (s *ResetStore) SaveToken(ctx context.Context, token, userID string, resetTokenTTL time.Duration) error {
	return s.client.SetEx(
		ctx,
		s.tokenKey(token),
		userID,
		resetTokenTTL,
	).Err()
}

func (s *ResetStore) GetToken(ctx context.Context, token string) (string, error) {
	return s.client.Get(ctx, s.tokenKey(token)).Result()
}

func (s *ResetStore) DeleteCode(ctx context.Context, userID string) error {
	key := s.codeKey(userID)

	if err := s.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete reset code: %w", err)
	}

	return nil
}

func (s *ResetStore) DeleteToken(ctx context.Context, token string) error {
	key := s.tokenKey(token)

	if err := s.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete reset token: %w", err)
	}

	return nil
}
