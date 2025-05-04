package event

import (
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"treffly/api/common"
	eventdto "treffly/api/dto/event"
	"treffly/api/models"
	"treffly/apperror"
)

type queryService interface {
	GetHomeForUser(ctx context.Context, params models.GetHomeParams) (*models.HomeEvents, error)
	GetHomeForGuest(ctx context.Context, params models.GetHomeParams) (*models.HomeEvents, error)
	GetUpcomingUserEvents(ctx context.Context, userID int32) ([]models.EventWithImages, error)
	GetPastUserEvents(ctx context.Context, userID int32) ([]models.EventWithImages, error)
	GetOwnedUserEvents(ctx context.Context, userID int32) ([]models.EventWithImages, error)
}

type QueryHandler struct {
	BaseHandler
	queryService queryService
	imageService ImageService
	converter    *eventdto.EventConverter
}

func NewEventQueryHandler(queryService queryService, imageService ImageService, converter *eventdto.EventConverter) *QueryHandler {
	return &QueryHandler{
		queryService: queryService,
		imageService: imageService,
		converter:    converter,
	}
}

func (h *QueryHandler) GetHome(ctx *gin.Context) {
	userID := common.GetUserIDFromSoftAuth(ctx)
	lat, lon, err := common.GetUserLocation(ctx)
	if err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	var homeEvents *models.HomeEvents
	if userID > 0 {
		homeEvents, err = h.queryService.GetHomeForUser(ctx, models.GetHomeParams{
			UserID: userID,
			Lat:    lat,
			Lon:    lon,
		})
		if err != nil {
			ctx.Error(apperror.WrapDBError(err))
			return
		}
	} else {
		homeEvents, err = h.queryService.GetHomeForGuest(ctx, models.GetHomeParams{
			UserID: userID,
			Lat:    lat,
			Lon:    lon,
		})
		if err != nil {
			ctx.Error(apperror.WrapDBError(err))
			return
		}
	}

	resp := h.converter.ToHomeEventsResponse(homeEvents)

	ctx.JSON(http.StatusOK, resp)
}

func (h *QueryHandler) GetUpcoming(ctx *gin.Context) {
	userID := common.GetUserIDFromContextPayload(ctx)

	events, err := h.queryService.GetUpcomingUserEvents(ctx, userID)
	if err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}

	resp := h.converter.ToEventsWithImages(events)

	ctx.JSON(http.StatusOK, resp)
}

func (h *QueryHandler) GetPast(ctx *gin.Context) {
	userID := common.GetUserIDFromContextPayload(ctx)

	events, err := h.queryService.GetPastUserEvents(ctx, userID)
	if err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}

	resp := h.converter.ToEventsWithImages(events)

	ctx.JSON(http.StatusOK, resp)
}

func (h *QueryHandler) GetOwned(ctx *gin.Context) {
	userID := common.GetUserIDFromContextPayload(ctx)

	events, err := h.queryService.GetOwnedUserEvents(ctx, userID)
	if err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}

	resp := h.converter.ToEventsWithImages(events)

	ctx.JSON(http.StatusOK, resp)
}
