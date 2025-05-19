package api

import (
	"errors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"time"
	"treffly/api/common"
	"treffly/api/models"
	"treffly/apperror"
	"treffly/token"
)

const (
	authorizationPayloadKey = "authorization_payload"
)

func authMiddleware(tokenMaker token.Maker) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		accessToken, err := ctx.Cookie("access_token")
		if err != nil {
			if errors.Is(err, http.ErrNoCookie) {
				ctx.Error(apperror.TokenExpired.WithCause(err))
				ctx.Abort()
				return
			}
			ctx.Error(err)
			ctx.Abort()
			return
		}

		payload, err := tokenMaker.VerifyToken(accessToken)
		if err != nil {
			ctx.Error(apperror.TokenExpired.WithCause(err))
			ctx.Abort()
			return
		}

		ctx.Set(authorizationPayloadKey, payload)
		ctx.Next()
	}
}

func ErrorHandler(log *zap.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()

		if len(ctx.Errors) > 0 {
			err := ctx.Errors.Last()
			var e apperror.ErrorResponse
			switch {
			case errors.As(err.Err, &e):
				log.Error("Request error",
					zap.String("path", ctx.FullPath()),
					zap.String("method", ctx.Request.Method),
					zap.Int("status", ctx.Writer.Status()),
					zap.Error(e.Unwrap()),
				)
				ctx.JSON(e.HTTPCode, e)
			default:
				log.Error("Request error",
					zap.String("path", ctx.FullPath()),
					zap.String("method", ctx.Request.Method),
					zap.Int("status", ctx.Writer.Status()),
					zap.Error(err),
				)
				ctx.JSON(http.StatusInternalServerError, apperror.InternalServer)
			}
		}
	}
}

func softAuthMiddleware(tokenMaker token.Maker) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		accessToken, err := ctx.Cookie("access_token")
		if err != nil {
			ctx.Next()
			return
		}

		payload, err := tokenMaker.VerifyToken(accessToken)
		if err != nil {
			ctx.Next()
			return
		}

		userID := payload.UserID
		ctx.Set("user_id", userID)
		ctx.Next()
	}
}

type rateLimitStore interface {
	CheckDescriptionLimit(ctx *gin.Context, endpoint string, userID string, limit int, window time.Duration) (models.RateLimitResult, error)
}

func RateLimitMiddleware(store rateLimitStore, limit int, window time.Duration) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userID := common.GetUserIDFromContextPayload(ctx)
		endpoint := ctx.FullPath()

		result, err := store.CheckDescriptionLimit(ctx, endpoint, string(userID), limit, window)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, apperror.InternalServer.WithCause(err))
			return
		}

		ctx.Set("rate_limit", result)

		if !result.Allowed {
			ctx.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many requests",
				"rate_limit": map[string]interface{}{
					"reset_at": result.ResetAt,
				},
			})
			return
		}

		ctx.Next()
	}
}