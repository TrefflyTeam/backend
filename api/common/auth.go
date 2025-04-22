package common

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

const (
	AuthorizationPayloadKey = "authorization_payload"
	AccessTokenCookiePath  = "/"
	RefreshTokenCookiePath = "/auth"
	CookieDomain = ""
)

func SetTokenCookie(ctx *gin.Context, name, token, path string, maxAge time.Duration, environment string) {
	ctx.SetSameSite(http.SameSiteLaxMode)
	isSecure := false
	if environment == "production" {
		isSecure = true
		path = "/api"+path
	}
	ctx.SetCookie(
		name,
		token,
		int(maxAge.Seconds()),
		path,
		CookieDomain, //TODO: set domain before releasing
		isSecure, //TODO: set to true before releasing
		true,
	)
}