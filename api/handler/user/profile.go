package user

import (
	"context"
	"database/sql"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"treffly/api/common"
	userdto "treffly/api/dto/user"
	"treffly/api/models"
	"treffly/apperror"
)

type updater interface {
	UpdateUser(ctx context.Context, params models.UpdateUserParams) (models.UserWithTags, error)
	GetUserWithTags(ctx context.Context, userID int32) (models.UserWithTags, error)
}

type deleter interface {
	DeleteUser(ctx context.Context, userID int32) error
}

type tagManager interface {
	UpdateUserTags(ctx context.Context, params models.UpdateUserTagsParams) error
}

type ProfileHandler struct {
	updater      updater
	deleter      deleter
	tagManager   tagManager
	imageService imageService
	converter    *userdto.UserConverter
	env          string
}

func NewProfileHandler(
	updater updater,
	deleter deleter, tagManager tagManager,
	imageService imageService,
	converter *userdto.UserConverter,
	env string,
	) *ProfileHandler {
	return &ProfileHandler{
		updater:      updater,
		deleter:      deleter,
		tagManager:   tagManager,
		imageService: imageService,
		converter:    converter,
		env:          env,
	}
}

func (h *ProfileHandler) GetCurrent(ctx *gin.Context) {
	userID := common.GetUserIDFromContextPayload(ctx)

	user, err := h.updater.GetUserWithTags(ctx, userID)
	if err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}

	resp := h.converter.ToUserWithTagsResponse(user)

	ctx.JSON(http.StatusOK, resp)
}

func (h *ProfileHandler) UpdateCurrent(ctx *gin.Context) {
	userID := common.GetUserIDFromContextPayload(ctx)

	var req userdto.UpdateUserRequest
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	var (
		oldImageID uuid.UUID
		oldPath    string
		err        error
	)

	oldImageID, oldPath, err = h.imageService.GetDBImageByUserID(ctx, userID)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			ctx.Error(apperror.WrapDBError(err))
			return
		}
	}

	var (
		imageID = uuid.Nil
		path    string
	)

	file, header, err := ctx.Request.FormFile("image")
	if err != nil && !errors.Is(err, http.ErrMissingFile) {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	if err == nil && file != nil && !req.DeleteImage {
		imageID = uuid.New()
		path, err = h.imageService.Upload(file, header, "user", imageID.String())
		if err != nil {
			ctx.Error(apperror.BadRequest.WithCause(err))
			return
		}
	}

	user, err := h.updater.UpdateUser(ctx, models.UpdateUserParams{
		ID:          userID,
		Username:    req.Username,
		NewImageID:  imageID,
		Path:        path,
		OldImageID:  oldImageID,
		DeleteImage: req.DeleteImage,
	})
	if err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}

	if req.DeleteImage && oldPath != "" {
		err = h.imageService.Delete(oldPath) //TODO: make deletes transactional
		if err != nil {
			ctx.Error(apperror.WrapDBError(err))
			return
		}
	}

	if oldPath != "" && file != nil {
		_ = h.imageService.Delete(oldPath)
	}

	resp := h.converter.ToUserWithTagsResponse(user)

	ctx.JSON(http.StatusOK, resp)
}

func (h *ProfileHandler) DeleteCurrent(ctx *gin.Context) {
	userID := common.GetUserIDFromContextPayload(ctx)

	_, path, err := h.imageService.GetDBImageByUserID(ctx, userID) //TODO: make deletes transactional
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		ctx.Error(apperror.WrapDBError(err))
		return
	}

	if path != "" {
		err = h.imageService.Delete(path)
		if err != nil {
			ctx.Error(apperror.WrapDBError(err))
			return
		}
	}

	err = h.deleter.DeleteUser(ctx, userID)
	if err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}

	isSecure := false
	path = ""
	if h.env == "production" {
		isSecure = true
		path = "/api"
	}
	ctx.SetCookie("access_token", "", -1, path+common.AccessTokenCookiePath,
		common.CookieDomain, isSecure, true)
	ctx.SetCookie("refresh_token", "", -1, path+common.RefreshTokenCookiePath,
		common.CookieDomain, isSecure, true)

	ctx.Status(http.StatusNoContent)
}

func (h *ProfileHandler) UpdateCurrentTags(ctx *gin.Context) {
	userID := common.GetUserIDFromContextPayload(ctx)

	var req userdto.UpdateCurrentUserTagsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	if err := h.tagManager.UpdateUserTags(ctx, models.UpdateUserTagsParams{
		UserID: userID,
		TagIDs: req.TagIDs,
	}); err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{})
}
