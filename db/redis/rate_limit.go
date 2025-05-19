package redis

import (
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
	"treffly/api/models"
)

type RateLimitStore struct {
	client *Client
}

func NewRateLimitStore(client *Client) RateLimitStore {
	return RateLimitStore{client: client}
}

func (s *RateLimitStore) CheckDescriptionLimit(
	ctx context.Context,
	endpoint string,
	userID string,
	limit int,
	window time.Duration,
) (models.RateLimitResult, error) {
	key := fmt.Sprintf("rate_limit:%s:%s", endpoint, userID)

	pipe := s.client.Pipeline()
	incr := pipe.Incr(ctx, key)
	ttl := pipe.TTL(ctx, key)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return models.RateLimitResult{}, err
	}

	count := incr.Val()
	currentTTL := ttl.Val()

	if count == 1 {
		if err = s.client.Expire(ctx, key, window).Err(); err != nil {
			return models.RateLimitResult{}, err
		}
		currentTTL = window
	}

	resetAt := time.Now().Add(currentTTL)

	return models.RateLimitResult{
		Allowed:    count <= int64(limit),
		Remaining:  max(limit-int(count), 0),
		ResetAt:    resetAt,
	}, nil
}

func (s *RateLimitStore) GetRateLimitInfo(
	ctx context.Context,
	endpoint string,
	userID string,
	limit int,
	window time.Duration,
) (*models.RateLimitResult, error) {
	key := fmt.Sprintf("rate_limit:%s:%s", endpoint, userID)

	countStr, err := s.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return &models.RateLimitResult{
			Allowed:   true,
			Remaining: limit,
			ResetAt:   time.Now().Add(window),
		}, nil
	} else if err != nil {
		return nil, err
	}

	count, err := strconv.Atoi(countStr)
	if err != nil {
		return nil, err
	}

	ttl, err := s.client.TTL(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	resetAt := time.Now().Add(ttl)
	if ttl == -1 {
		resetAt = time.Now().Add(window)
	}

	return &models.RateLimitResult{
		Allowed:    count <= limit,
		Remaining:  max(limit-count, 0),
		ResetAt:    resetAt,
	}, nil
}

func (s *RateLimitStore) CanSendResetRequest(ctx context.Context, email string, rateLimit time.Duration) (bool, error) {
	key := "rate_limit:" + email
	now := time.Now().Unix()

	result, err := s.client.SetNX(ctx, key, now, rateLimit).Result()
	if err != nil {
		return false, err
	}

	if !result {
		lastSent, err := s.client.Get(ctx, key).Int64()
		if err != nil {
			return false, err
		}
		if now-lastSent < int64(rateLimit.Seconds()) {
			return false, nil
		}

		s.client.Set(ctx, key, now, rateLimit)
	}

	return true, nil
}
