package handler

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"strconv"
	"strings"
	"treffly/api/common"
	"treffly/api/dto/event"
	eventservice "treffly/api/service/event"
	imageservice "treffly/api/service/image"
	"treffly/apperror"
	"treffly/util"
)

type EventHandler struct {
	eventService *eventservice.Service
	imageService *imageservice.Service
	config       util.Config
}

func NewEventHandler(eventService *eventservice.Service, imageService *imageservice.Service, config util.Config) *EventHandler {
	return &EventHandler{
		eventService: eventService,
		imageService: imageService,
		config:       config,
	}
}

func (h *EventHandler) Create(ctx *gin.Context) {
	userID := common.GetUserIDFromContextPayload(ctx)

	var req eventdto.CreateEventRequest
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	var (
		imageID uuid.UUID
		path    string
	)

	file, header, err := ctx.Request.FormFile("image")
	if err != nil && !errors.Is(err, http.ErrMissingFile) {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	if err == nil {
		imageID = uuid.New()
		path, err = h.imageService.Upload(file, header, "event", imageID.String())
	}
	params := eventservice.CreateParams{
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
		ImageID:     imageID,
		Path:        path,
	}

	createdEvent, err := h.eventService.Create(ctx, params)
	if err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}

	imageEventURL := imageservice.ImageURL(h.config.Environment, h.config.Domain, path)
	imageUserURL := imageservice.ImageURL(h.config.Environment, h.config.Domain, createdEvent.GetUserImagePath())

	resp := eventdto.NewCreateEventResponse(eventdto.ConvertEvent(createdEvent), imageEventURL, imageUserURL)

	ctx.JSON(http.StatusOK, resp)
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

	params := eventservice.ListParams{
		Lat:       lat,
		Lon:       lon,
		Search:    ctx.Query("keywords"),
		TagIDs:    tagIDs,
		DateRange: ctx.Query("dateWithin"),
	}

	events, err := h.eventService.List(ctx, params)
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

	eventWithStatus, err := h.eventService.GetEventWithStatus(ctx, int32(eventID), userID)
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

	params := eventservice.UpdateParams{
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

	updatedEvent, err := h.eventService.Update(ctx, params)
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

	err = h.eventService.Delete(ctx, eventservice.DeleteParams{
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

	params := eventservice.SubscriptionParams{
		EventID: int32(eventID),
		UserID:  userID,
	}

	var result eventservice.EventWithStatus
	if subscribe {
		result, err = h.eventService.Subscribe(ctx, params)
	} else {
		result, err = h.eventService.Unsubscribe(ctx, params)
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

	homeData, err := h.eventService.GetHome(ctx, eventservice.GetHomeParams{
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

	events, err := h.eventService.GetUpcomingUserEvents(ctx, userID)
	if err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}

	ctx.JSON(http.StatusOK, eventdto.ConvertEvents(events))
}

func (h *EventHandler) GetPast(ctx *gin.Context) {
	userID := common.GetUserIDFromContextPayload(ctx)

	events, err := h.eventService.GetPastUserEvents(ctx, userID)
	if err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}

	ctx.JSON(http.StatusOK, eventdto.ConvertEvents(events))
}

func (h *EventHandler) GetOwned(ctx *gin.Context) {
	userID := common.GetUserIDFromContextPayload(ctx)

	events, err := h.eventService.GetOwnedUserEvents(ctx, userID)
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
