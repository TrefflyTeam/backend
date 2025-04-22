package handler

import (
	"database/sql"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"treffly/api/common"
	tokenservice "treffly/api/service/token"
	"treffly/apperror"
	"treffly/util"
)

type TokenHandler struct {
	service *tokenservice.Service
	config util.Config
}

func NewTokenHandler(service *tokenservice.Service, config util.Config) *TokenHandler {
	return &TokenHandler{
		service: service,
		config: config,
	}
}

func (h *TokenHandler) RefreshTokens(ctx *gin.Context) {
	reqRefreshToken, err := ctx.Cookie("refresh_token")
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			ctx.Error(apperror.TokenExpired.WithCause(err))
			return
		}
		ctx.Error(err)
		return
	}

	accessToken, refreshToken, err := h.service.RefreshTokens(ctx, reqRefreshToken)
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
