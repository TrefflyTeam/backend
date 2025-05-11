package redis

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"time"
	"treffly/api/models"
)

type RateLimitStore struct {
	client *Client
}

func NewRateLimitStore(client *Client) *RateLimitStore {
	return &RateLimitStore{client: client}
}

func (s *RateLimitStore) CheckDescriptionLimit(
	ctx *gin.Context,
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
