package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
	"treffly/api/common"
	"treffly/api/dto/event"
	userservice "treffly/api/service/event"
	"treffly/apperror"
)

type EventHandler struct {
	service *userservice.Service
}

func NewEventHandler(service *userservice.Service) *EventHandler {
	return &EventHandler{service: service}
}

func (h *EventHandler) Create(ctx *gin.Context) {
	userID := common.GetUserIDFromContextPayload(ctx)

	var req eventdto.CreateEventRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	params := userservice.CreateParams{
		Name:        req.Name,
		Description: req.Description,
		Capacity:    req.Capacity,
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
		Address:     req.Address,
		Date:        req.Date,
		IsPrivate:   req.IsPrivate,
		Tags:        req.Tags,
		OwnerID:     userID,
	}

	createdEvent, err := h.service.Create(ctx, params)
	if err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}

	ctx.JSON(http.StatusOK, eventdto.ConvertEvent(createdEvent))
}

func (h *EventHandler) List(ctx *gin.Context) {
	lat, lon, err := common.GetUserLocation(ctx)
	if err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	tagIDs, err := parseTagIDs(ctx.Query("tags"))
	if err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	params := userservice.ListParams{
		Lat:       lat,
		Lon:       lon,
		Search:    ctx.Query("keywords"),
		TagIDs:    tagIDs,
		DateRange: ctx.Query("dateWithin"),
	}

	events, err := h.service.List(ctx, params)
	if err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}

	ctx.JSON(http.StatusOK, eventdto.ConvertEvents(events))
}

func (h *EventHandler) GetByID(ctx *gin.Context) {
	eventID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	userID := common.GetUserIDFromSoftAuth(ctx)

	eventWithStatus, err := h.service.GetEventWithStatus(ctx, int32(eventID), userID)
	if err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}

	resp := eventdto.NewEventByIDResponse(eventdto.ConvertEvent(eventWithStatus.Event),
		eventWithStatus.IsOwner, eventWithStatus.IsParticipant)

	ctx.JSON(http.StatusOK, resp)
}

func (h *EventHandler) Update(ctx *gin.Context) {
	userID := common.GetUserIDFromContextPayload(ctx)
	eventID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	var req eventdto.UpdateEventRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	params := userservice.UpdateParams{
		EventID:     int32(eventID),
		Name:        req.Name,
		Description: req.Description,
		Capacity:    req.Capacity,
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
		Address:     req.Address,
		Date:        req.Date,
		IsPrivate:   req.IsPrivate,
		Tags:        req.Tags,
		UserID:      userID,
	}

	updatedEvent, err := h.service.Update(ctx, params)
	if err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}

	ctx.JSON(http.StatusOK, eventdto.ConvertEvent(updatedEvent))
}

func (h *EventHandler) Delete(ctx *gin.Context) {
	userID := common.GetUserIDFromContextPayload(ctx)
	eventID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	err = h.service.Delete(ctx, userservice.DeleteParams{
		EventID: int32(eventID),
		UserID:  userID,
	})
	if err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (h *EventHandler) Subscribe(ctx *gin.Context) {
	h.handleSubscription(ctx, true)
}

func (h *EventHandler) Unsubscribe(ctx *gin.Context) {
	h.handleSubscription(ctx, false)
}

func (h *EventHandler) handleSubscription(ctx *gin.Context, subscribe bool) {
	userID := common.GetUserIDFromContextPayload(ctx)
	eventID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	params := userservice.SubscriptionParams{
		EventID: int32(eventID),
		UserID:  userID,
	}

	var result userservice.EventWithStatus
	if subscribe {
		result, err = h.service.Subscribe(ctx, params)
	} else {
		result, err = h.service.Unsubscribe(ctx, params)
	}

	if err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}

	resp := eventdto.NewEventByIDResponse(eventdto.ConvertEvent(result.Event), result.IsOwner, result.IsParticipant)
	
	ctx.JSON(http.StatusOK, resp)
}

func (h *EventHandler) GetHome(ctx *gin.Context) {
	userID := common.GetUserIDFromSoftAuth(ctx)
	lat, lon, err := common.GetUserLocation(ctx)
	if err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	homeData, err := h.service.GetHome(ctx, userservice.GetHomeParams{
		UserID: userID,
		Lat:    lat,
		Lon:    lon,
	})
	if err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}

	resp := eventdto.NewGetHomeEventsResponse(homeData.Premium, homeData.Recommended, homeData.Latest, homeData.Popular)
	
	ctx.JSON(http.StatusOK, resp)
}

func (h *EventHandler) GetUpcoming(ctx *gin.Context) {
	userID := common.GetUserIDFromContextPayload(ctx)

	events, err := h.service.GetUpcomingUserEvents(ctx, userID)
	if err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}

	ctx.JSON(http.StatusOK, eventdto.ConvertEvents(events))
}

func (h *EventHandler) GetPast(ctx *gin.Context) {
	userID := common.GetUserIDFromContextPayload(ctx)

	events, err := h.service.GetPastUserEvents(ctx, userID)
	if err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}

	ctx.JSON(http.StatusOK, eventdto.ConvertEvents(events))
}

func (h *EventHandler) GetOwned(ctx *gin.Context) {
	userID := common.GetUserIDFromContextPayload(ctx)

	events, err := h.service.GetOwnedUserEvents(ctx, userID)
	if err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}

	ctx.JSON(http.StatusOK, eventdto.ConvertEvents(events))
}


func parseTagIDs(tagsStr string) ([]int32, error) {
	if tagsStr == "" {
		return []int32{}, nil
	}

	strIDs := strings.Split(tagsStr, ",")
	result := make([]int32, 0, len(strIDs))

	for _, strID := range strIDs {
		cleaned := strings.TrimSpace(strID)
		if cleaned == "" {
			continue
		}

		id, err := strconv.ParseInt(cleaned, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid tag ID: %s", strID)
		}

		result = append(result, int32(id))
	}

	return result, nil
}