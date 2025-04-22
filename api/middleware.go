package api

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"treffly/api/common"
	"treffly/apperror"
	"treffly/token"
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

		ctx.Set(common.AuthorizationPayloadKey, payload)
		ctx.Next()
	}
}

func ErrorHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()

		if len(ctx.Errors) > 0 {
			err := ctx.Errors.Last()
			var e apperror.ErrorResponse
			switch {
			case errors.As(err.Err, &e):
				ctx.JSON(e.HTTPCode, e)
			default:
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
