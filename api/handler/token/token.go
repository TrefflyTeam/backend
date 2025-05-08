package token

import (
	"context"
	"database/sql"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"treffly/api/common"
	"treffly/apperror"
	"treffly/util"
)

type tokenManager interface {
	RefreshTokens(ctx context.Context, reqRefreshToken string) (accessToken string, refreshToken string, err error)
	ValidateSession(ctx context.Context, refreshToken string) error
}

type Handler struct {
	tokenManager tokenManager
	config util.Config
}

func NewTokenHandler(tokenManager tokenManager, config util.Config) *Handler {
	return &Handler{
		tokenManager: tokenManager,
		config: config,
	}
}

func (h *Handler) RefreshTokens(ctx *gin.Context) {
	reqRefreshToken, err := ctx.Cookie("refresh_token")
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			ctx.Error(apperror.TokenExpired.WithCause(err))
			return
		}
		ctx.Error(err)
		return
	}

	accessToken, refreshToken, err := h.tokenManager.RefreshTokens(ctx, reqRefreshToken)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ctx.Error(apperror.NotFound.WithCause(err))
			return
		}

		ctx.Error(apperror.WrapDBError(err))
		return
	}


	common.SetTokenCookie(ctx, "access_token", accessToken,
		common.AccessTokenCookiePath, h.config.AccessTokenDuration, h.config.Environment)
	common.SetTokenCookie(ctx, "refresh_token", refreshToken,
		common.RefreshTokenCookiePath, h.config.RefreshTokenDuration, h.config.Environment)

	ctx.JSON(http.StatusOK, gin.H{})
}

func (h *Handler) Auth(ctx *gin.Context) {
	token, err := ctx.Cookie("refresh_token")
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{})
		return
	}
	err = h.tokenManager.ValidateSession(ctx, token)
	if err != nil {
		common.SetTokenCookie(ctx, "refresh_token", "",
			common.RefreshTokenCookiePath, -1, h.config.Environment)
		ctx.Error(apperror.TokenExpired.WithCause(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{})
}
