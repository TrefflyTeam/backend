package user

import (
	"context"
	"database/sql"
	"errors"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"treffly/api/common"
	userdto "treffly/api/dto/user"
	"treffly/api/models"
	"treffly/apperror"
	"treffly/util"
)

type creator interface {
	CreateUser(ctx context.Context, params models.CreateUserParams) (models.User, error)
}

type authService interface {
	LoginUser(ctx context.Context, email, password string) (models.User, string, string, error)
	CreateAuthSession(ctx context.Context, userID int32) (string, string, error)
}

type AuthHandler struct {
	authService authService
	creator     creator
	converter   *userdto.UserConverter
	config      util.Config
}

func NewAuthHandler(authService authService, creator creator, converter *userdto.UserConverter, config util.Config) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		creator:     creator,
		converter:   converter,
		config:      config,
	}
}

func (h *AuthHandler) Create(ctx *gin.Context) {
	var req userdto.CreateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	user, err := h.creator.CreateUser(ctx, models.CreateUserParams{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}

	accessToken, refreshToken, err := h.authService.CreateAuthSession(ctx, user.ID)
	if err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}

	h.setAuthCookies(ctx, accessToken, refreshToken)

	resp := h.converter.ToUserResponse(user)

	ctx.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) Login(ctx *gin.Context) {
	var req userdto.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	user, accessToken, refreshToken, err := h.authService.LoginUser(ctx, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ctx.Error(apperror.InvalidCredentials.WithCause(err))
			return
		}
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			ctx.Error(apperror.InvalidCredentials.WithCause(err))
			return
		}
		ctx.Error(apperror.WrapDBError(err))
		return
	}

	h.setAuthCookies(ctx, accessToken, refreshToken)

	resp := h.converter.ToUserResponse(user)

	ctx.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) setAuthCookies(ctx *gin.Context, accessToken, refreshToken string) {
	common.SetTokenCookie(ctx, "access_token", accessToken,
		common.AccessTokenCookiePath, h.config.AccessTokenDuration, h.config.Environment)
	common.SetTokenCookie(ctx, "refresh_token", refreshToken,
		common.RefreshTokenCookiePath, h.config.RefreshTokenDuration, h.config.Environment)
}

func (h *AuthHandler) Logout(ctx *gin.Context) {
	isSecure := false
	path := "" //TODO: define path vars on server init
	if h.config.Environment == "production" {
		isSecure = true
		path = "/api"
	}
	ctx.SetCookie("access_token", "", -1, path+common.AccessTokenCookiePath,
		common.CookieDomain, isSecure, true)
	ctx.SetCookie("refresh_token", "", -1, path+common.RefreshTokenCookiePath,
		common.CookieDomain, isSecure, true)
	//TODO: block session
	ctx.JSON(http.StatusNoContent, gin.H{})
	ctx.Status(http.StatusNoContent)
}
