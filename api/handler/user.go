package handler

import (
	"net/http"
	"treffly/api/common"
	userdto "treffly/api/dto/user"
	imageservice "treffly/api/service/image"
	userservice "treffly/api/service/user"
	"treffly/apperror"
	"treffly/util"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	service      *userservice.Service
	imageService *imageservice.Service
	config       util.Config
}

func NewUserHandler(service *userservice.Service, imageService *imageservice.Service, config util.Config) *UserHandler {
	return &UserHandler{
		service:      service,
		imageService: imageService,
		config:       config,
	}
}

func (h *UserHandler) Create(ctx *gin.Context) {
	var req userdto.CreateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	user, err := h.service.CreateUser(ctx, userservice.CreateParams{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}

	accessToken, refreshToken, err := h.service.CreateAuthSession(ctx, user.ID)
	if err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}

	h.setAuthCookies(ctx, accessToken, refreshToken)
	ctx.JSON(http.StatusOK, userdto.NewLoginResponse(user))
}

func (h *UserHandler) Login(ctx *gin.Context) {
	var req userdto.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	user, accessToken, refreshToken, err := h.service.LoginUser(ctx, req.Email, req.Password)
	if err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}

	h.setAuthCookies(ctx, accessToken, refreshToken)
	ctx.JSON(http.StatusOK, userdto.NewLoginResponse(user))
}

func (h *UserHandler) setAuthCookies(ctx *gin.Context, accessToken, refreshToken string) {
	common.SetTokenCookie(ctx, "access_token", accessToken,
		common.AccessTokenCookiePath, h.config.AccessTokenDuration, h.config.Environment)
	common.SetTokenCookie(ctx, "refresh_token", refreshToken,
		common.RefreshTokenCookiePath, h.config.RefreshTokenDuration, h.config.Environment)
}

func (h *UserHandler) Logout(ctx *gin.Context) {
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

func (h *UserHandler) GetCurrent(ctx *gin.Context) {
	userID := common.GetUserIDFromContextPayload(ctx)

	user, err := h.service.GetUserWithTags(ctx, userID)
	if err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}

	ctx.JSON(http.StatusOK, userdto.NewUserWithTagsResponse(user))
}

func (h *UserHandler) UpdateCurrent(ctx *gin.Context) {
	userID := common.GetUserIDFromContextPayload(ctx)

	var req userdto.UpdateUserRequest
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	user, err := h.service.UpdateUser(ctx, userservice.UpdateUserParams{
		ID:       userID,
		Username: req.Username,
	})

	if err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}

	//id := uuid.New()
	//url, err := h.imageService.Upload(ctx, "users", id.String())
	//if err != nil {
	//	ctx.Error(apperror.WrapDBError(err))
	//	return
	//}

	ctx.JSON(http.StatusOK, userdto.NewUpdateUserResponse(user, ""))
}

func (h *UserHandler) DeleteCurrent(ctx *gin.Context) {
	userID := common.GetUserIDFromContextPayload(ctx)

	if err := h.service.DeleteUser(ctx, userID); err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}

	h.Logout(ctx)

	ctx.Status(http.StatusNoContent)
}

func (h *UserHandler) UpdateCurrentTags(ctx *gin.Context) {
	userID := common.GetUserIDFromContextPayload(ctx)

	var req userdto.UpdateCurrentUserTagsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	if err := h.service.UpdateUserTags(ctx, userservice.UpdateUserTagsParams{
		UserID: userID,
		TagIDs: req.TagIDs,
	}); err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{})
}

func (h *UserHandler) Auth(ctx *gin.Context) {
	_, err := ctx.Cookie("refresh_token")
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{})
}
