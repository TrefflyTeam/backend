package event

import (
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	eventdto "treffly/api/dto/event"
	"treffly/api/models"
	"treffly/apperror"
)

type adminCrudService interface {
	List(ctx context.Context, params models.ListParams) ([]models.Event, error)
	AdminDelete(ctx context.Context, id int32) error
}

type AdminCRUDHandler struct {
	BaseHandler
	crudService  adminCrudService
	imageService ImageService
	converter    *eventdto.EventConverter
}

func NewAdminEventCRUDHandler(crudService adminCrudService, imageService ImageService, converter *eventdto.EventConverter) *AdminCRUDHandler {
	return &AdminCRUDHandler{crudService: crudService, imageService: imageService, converter: converter}
}

func (h *AdminCRUDHandler) List(ctx *gin.Context) {
	tagIDs, err := parseTagIDs(ctx.Query("tags"))
	if err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	params := models.ListParams{
		Search:    ctx.Query("keywords"),
		TagIDs:    tagIDs,
		DateRange: ctx.Query("dateWithin"),
	}

	events, err := h.crudService.List(ctx, params)
	if err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}

	resp := h.converter.ToEventsResponse(events)

	ctx.JSON(http.StatusOK, resp)
}

func (h *AdminCRUDHandler) Delete(ctx *gin.Context) {
	eventID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	_, path, err := h.imageService.GetDBImageByEventID(ctx, int32(eventID)) //TODO: make deletes transactional
	if err != nil {
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

	err = h.crudService.AdminDelete(ctx, int32(eventID))
	if err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}

	ctx.Status(http.StatusNoContent)
}
