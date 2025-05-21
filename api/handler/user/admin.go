package user

import (
	"context"
	"database/sql"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	userdto "treffly/api/dto/user"
	"treffly/api/models"
	"treffly/apperror"
)

type adminCrudService interface {
	ListAll(ctx context.Context, username string) ([]models.User, error)
	AdminDelete(ctx context.Context, id int32) error
}

type AdminUserHandler struct {
	service      adminCrudService
	converter    *userdto.UserConverter
	imageService imageService
}

func NewAdminUserHandler(service adminCrudService, converter *userdto.UserConverter, imageService imageService) *AdminUserHandler {
	return &AdminUserHandler{service: service, converter: converter, imageService: imageService}
}

func (h *AdminUserHandler) ListAll(ctx *gin.Context) {
	username := ctx.Query("username")

	users, err := h.service.ListAll(ctx, username)
	if err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	var resp []userdto.AdminUserResponse
	for _, u := range users {
		resp = append(resp, h.converter.ToAdminUserResponse(u))
	}

	ctx.JSON(http.StatusOK, resp)
}

func (h *AdminUserHandler) Delete(ctx *gin.Context) {
	userID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	_, path, err := h.imageService.GetDBImageByUserID(ctx, int32(userID)) //TODO: make deletes transactional
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

	err = h.service.AdminDelete(ctx, int32(userID))
	if err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}

	ctx.Status(http.StatusNoContent)
}
