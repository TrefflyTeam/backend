package api

import (
	"database/sql"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
	db "treffly/db/sqlc"
	"treffly/util"
)

type CreateUserRequest struct {
	Username string `json:"username" binding:"required,username,min=2,max=16"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type userResponse struct {
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

func newUserResponse(user db.User) userResponse {
	return userResponse{
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}
}

func (server *Server) createUser(ctx *gin.Context) {
	var req CreateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	arg := db.CreateUserParams{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
	}

	user, err := server.store.CreateUser(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	resp := newUserResponse(user)

	ctx.JSON(http.StatusOK, resp)
}

type loginUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6,max=32"`
}

type loginUserResponse struct {
	User                  userResponse `json:"user"`
	AccessTokenExpiresAt  time.Time    `json:"access_token_expires_at"`
	RefreshTokenExpiresAt time.Time    `json:"refresh_token_expires_at"`
}

func (server *Server) loginUser(ctx *gin.Context) {
	var req loginUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	user, err := server.store.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	err = util.CheckPassword(req.Password, user.PasswordHash)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	accessToken, accessPayload, err := server.tokenMaker.CreateToken(
		user.ID,
		server.config.AccessTokenDuration,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	refreshToken, refreshPayload, err := server.tokenMaker.CreateToken(
		user.ID,
		server.config.RefreshTokenDuration,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	_, err = server.store.CreateSession(ctx, db.CreateSessionParams{
		Uuid:         refreshPayload.ID,
		UserID:       user.ID,
		RefreshToken: refreshToken,
		ExpiresAt:    refreshPayload.ExpiredAt,
		IsBlocked:    false,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	resp := loginUserResponse{
		User:                  newUserResponse(user),
		AccessTokenExpiresAt:  accessPayload.ExpiredAt,
		RefreshTokenExpiresAt: refreshPayload.ExpiredAt,
	}

	server.setTokenCookie(ctx, "access_token", accessToken, accessTokenCookiePath, server.config.AccessTokenDuration)
	server.setTokenCookie(ctx, "refresh_token", refreshToken, refreshTokenCookiePath, server.config.RefreshTokenDuration)

	ctx.JSON(http.StatusOK, resp)
}

func (server *Server) logoutUser(ctx *gin.Context) {
	isSecure := server.config.Environment == "production"
	ctx.SetCookie("access_token", "", -1, accessTokenCookiePath, cookieDomain, isSecure, true)
	ctx.SetCookie("refresh_token", "", -1, refreshTokenCookiePath, cookieDomain, isSecure, true)
	//TODO: block session
	ctx.JSON(http.StatusNoContent, gin.H{})
}
