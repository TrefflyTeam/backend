package api

import (
	"database/sql"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"time"
	"treffly/apperror"
	db "treffly/db/sqlc"
	"treffly/token"
	"treffly/util"
)

type CreateUserRequest struct {
	Username string `json:"username" binding:"required,username,min=2,max=20"`
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

type userWithTagsResponse struct {
	ID        int32     `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Tags      []db.Tag  `json:"tags"`
	CreatedAt time.Time `json:"created_at"`
}

func newUserWithTagsResponse(user db.UserWithTagsView) userWithTagsResponse {
	return userWithTagsResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Tags:      user.Tags,
		CreatedAt: user.CreatedAt,
	}
}

func (server *Server) createUser(ctx *gin.Context) {
	var req CreateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		ctx.Error(err)
		return
	}

	arg := db.CreateUserParams{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
	}

	user, err := server.store.CreateUser(ctx, arg)
	if err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}

	loginResp := server.createAuthSession(ctx, user)

	ctx.JSON(http.StatusOK, loginResp)
}

type loginUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type loginUserResponse struct {
	userResponse
}

func (server *Server) loginUser(ctx *gin.Context) {
	var req loginUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	user, err := server.store.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ctx.Error(apperror.BadRequest.WithCause(err))
			return
		}
		ctx.Error(apperror.WrapDBError(err))
		return
	}

	err = util.CheckPassword(req.Password, user.PasswordHash)
	if err != nil {
		ctx.Error(apperror.InvalidCredentials.WithCause(err))
		return
	}

	resp := server.createAuthSession(ctx, user)

	ctx.JSON(http.StatusOK, resp)
}

func (server *Server) createAuthSession(ctx *gin.Context, user db.User) (resp loginUserResponse) {
	accessToken, _, err := server.tokenMaker.CreateToken(
		user.ID,
		server.config.AccessTokenDuration,
	)
	if err != nil {
		ctx.Error(err)
		return
	}

	refreshToken, refreshPayload, err := server.tokenMaker.CreateToken(
		user.ID,
		server.config.RefreshTokenDuration,
	)
	if err != nil {
		ctx.Error(err)
		return
	}

	err = server.store.CreateSession(ctx, db.CreateSessionParams{
		Uuid:         refreshPayload.ID,
		UserID:       user.ID,
		RefreshToken: refreshToken,
		ExpiresAt:    refreshPayload.ExpiredAt,
		IsBlocked:    false,
	})
	if err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}

	resp = loginUserResponse{
		newUserResponse(user),
	}

	server.setTokenCookie(ctx, "access_token", accessToken, accessTokenCookiePath, server.config.AccessTokenDuration)
	server.setTokenCookie(ctx, "refresh_token", refreshToken, refreshTokenCookiePath, server.config.RefreshTokenDuration)

	return resp
}

func (server *Server) logoutUser(ctx *gin.Context) {
	isSecure := false
	path := "" //TODO: define path vars on server init
	if server.config.Environment == "production" {
		isSecure = true
		path = "/api"
	}
	ctx.SetCookie("access_token", "", -1, path+accessTokenCookiePath, cookieDomain, isSecure, true)
	ctx.SetCookie("refresh_token", "", -1, path+refreshTokenCookiePath, cookieDomain, isSecure, true)
	//TODO: block session
	ctx.JSON(http.StatusNoContent, gin.H{})
}

func (server *Server) auth(ctx *gin.Context) {
	_, err := ctx.Cookie("refresh_token")
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{})
}

func (server *Server) getCurrentUser(ctx *gin.Context) {
	userID := getUserIDFromContextPayload(ctx)

	user, err := server.store.GetUserWithTags(ctx, userID)
	if err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}

	ctx.JSON(http.StatusOK, newUserWithTagsResponse(user))
}

type updateUserRequest struct {
	Username string `json:"username" binding:"required,username,min=2,max=20"`
}

func (server *Server) updateCurrentUser(ctx *gin.Context) {
	userID := getUserIDFromContextPayload(ctx)

	var req updateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	arg := db.UpdateUserParams{
		ID: userID,
		Username: req.Username,
	}

	user, err := server.store.UpdateUser(ctx, arg)
	if err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}

	ctx.JSON(http.StatusOK, newUserResponse(user))
}

func (server *Server) deleteCurrentUser(ctx *gin.Context) {
	userID := getUserIDFromContextPayload(ctx)

	err := server.store.DeleteUser(ctx, userID)
	if err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}

	ctx.JSON(http.StatusNoContent, gin.H{})
}

type addTagResponse struct {
	UserID int32 `json:"user_id"`
	TagID  int32 `json:"tag_id"`
}

func newAddTagResponse(tag db.UserTag) addTagResponse {
	return addTagResponse{
		UserID: tag.UserID,
		TagID:  tag.TagID,
	}
}

func (server *Server) addCurrentUserTag(ctx *gin.Context) {
	userID := getUserIDFromContextPayload(ctx)

	id := ctx.Param("id")

	tagID, err := strconv.Atoi(id)
	if err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	arg := db.AddUserTagParams{
		UserID: userID,
		TagID: int32(tagID),
	}

	userTag, err := server.store.AddUserTag(ctx, arg)
	if err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}

	ctx.JSON(http.StatusOK, newAddTagResponse(userTag))
}

func (server *Server) deleteCurrentUserTag(ctx *gin.Context) {
	userID := getUserIDFromContextPayload(ctx)

	id := ctx.Param("id")

	tagID, err := strconv.Atoi(id)
	if err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	arg := db.DeleteUserTagParams{
		UserID: userID,
		TagID: int32(tagID),
	}

	err = server.store.DeleteUserTag(ctx, arg)
	if err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}

	ctx.JSON(http.StatusNoContent, gin.H{})
}

func getUserIDFromContextPayload(ctx *gin.Context) int32 {
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	userID := authPayload.UserID
	return userID
}