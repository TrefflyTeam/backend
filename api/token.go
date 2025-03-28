package api

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
	"treffly/apperror"
	db "treffly/db/sqlc"
)

const (
	accessTokenCookiePath  = "/"
	refreshTokenCookiePath = "/tokens/refresh"
	cookieDomain = ""
)

type refreshTokensResponse struct {
	AccessTokenExpiresAt time.Time `json:"access_token_expires_at"`
	RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at"`
}

func (server *Server) refreshTokens(ctx *gin.Context) {
	reqRefreshToken, err := ctx.Cookie("refresh_token")
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			ctx.Error(apperror.TokenExpired.WithCause(err))
			return
		}
		ctx.Error(err)
		return
	}

	reqRefreshPayload, err := server.tokenMaker.VerifyToken(reqRefreshToken)
	if err != nil {
		ctx.Error(apperror.TokenExpired.WithCause(err))
		return
	}

	session, err := server.store.GetSession(ctx, reqRefreshPayload.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ctx.Error(apperror.NotFound.WithCause(err))
			return
		}
		ctx.Error(err)
		return
	}

	if session.IsBlocked {
		err := fmt.Errorf("blocked session")
		ctx.Error(apperror.TokenExpired.WithCause(err))
		return
	}

	if session.UserID != reqRefreshPayload.UserID {
		err := fmt.Errorf("incorrect session user")
		ctx.Error(apperror.TokenExpired.WithCause(err))
		return
	}

	if session.RefreshToken != reqRefreshToken {
		err := fmt.Errorf("mismatched session token")
		ctx.Error(apperror.TokenExpired.WithCause(err))

		return
	}

	if time.Now().After(session.ExpiresAt) {
		err := fmt.Errorf("expired session")
		ctx.Error(apperror.TokenExpired.WithCause(err))
		return
	}

	accessToken, accessPayload, err := server.tokenMaker.CreateToken(
		reqRefreshPayload.UserID,
		server.config.AccessTokenDuration,
	)
	if err != nil {
		ctx.Error(apperror.InternalServer.WithCause(err))
		return
	}

	refreshToken, refreshPayload, err := server.tokenMaker.CreateToken(
		reqRefreshPayload.UserID,
		server.config.RefreshTokenDuration,
	)
	if err != nil {
		ctx.Error(apperror.InternalServer.WithCause(err))
		return
	}

	_, err = server.store.UpdateSession(ctx, db.UpdateSessionParams{
		OldUuid: reqRefreshPayload.ID,
		NewUuid: refreshPayload.ID,
		RefreshToken: refreshToken,
		ExpiresAt: refreshPayload.ExpiredAt,
	})
	if err != nil {
		ctx.Error(apperror.InternalServer.WithCause(err))
		return
	}

	server.setTokenCookie(ctx, "access_token", accessToken, accessTokenCookiePath, server.config.AccessTokenDuration)
	server.setTokenCookie(ctx, "refresh_token", refreshToken, refreshTokenCookiePath, server.config.RefreshTokenDuration)

	rsp := refreshTokensResponse{
		AccessTokenExpiresAt: accessPayload.ExpiredAt,
		RefreshTokenExpiresAt: refreshPayload.ExpiredAt,
	}
	ctx.JSON(http.StatusOK, rsp)
}

func (server *Server) setTokenCookie(ctx *gin.Context, name, token, path string, maxAge time.Duration) {
	ctx.SetSameSite(http.SameSiteLaxMode)
	isSecure := false
	if server.config.Environment == "production" {
		isSecure = true
		path = "/api"+path
	}
	ctx.SetCookie(
		name,
		token,
		int(maxAge.Seconds()),
		path,
		cookieDomain, //TODO: set domain before releasing
		isSecure, //TODO: set to true before releasing
		true,
	)
}

