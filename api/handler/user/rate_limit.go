package user

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
	"treffly/api/common"
	"treffly/api/models"
	"treffly/apperror"
)

type limitChecker interface {
	GetRateLimitInfo(ctx *gin.Context, endpoint string, userID string, limit int, window time.Duration) (*models.RateLimitResult, error)
}

type LimitCheckHandler struct {
	limitChecker limitChecker
	limit        int
	window       time.Duration
}

func NewLimitCheckHandler(limitChecker limitChecker, limit int, window time.Duration) *LimitCheckHandler {
	return &LimitCheckHandler{
		limitChecker: limitChecker,
		limit:        limit,
		window:       window,
	}
}

func (g *LimitCheckHandler) CheckGenerateRateLimit(ctx *gin.Context) {
	userID := common.GetUserIDFromContextPayload(ctx)

	result, err := g.limitChecker.GetRateLimitInfo(ctx, "/events/generate-desc", string(userID), g.limit, g.window)
	if err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"limit":     g.limit,
		"remaining": result.Remaining,
		"reset_at":  result.ResetAt.String(),
	})
}
