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

type subscriptionService interface {
	Subscribe(ctx context.Context, params models.SubscriptionParams) (models.Event, error)
	Unsubscribe(ctx context.Context, params models.SubscriptionParams) (models.Event, error)
}

type SubscriptionHandler struct {
	BaseHandler
	service subscriptionService
	converter *eventdto.EventConverter
}

func NewEventSubscriptionHandler(service subscriptionService, converter *eventdto.EventConverter) *SubscriptionHandler {
	return &SubscriptionHandler{
		service: service,
		converter: converter,
	}
}

func (h *SubscriptionHandler) Subscribe(ctx *gin.Context) {
	h.handleSubscription(ctx, true)
}

func (h *SubscriptionHandler) Unsubscribe(ctx *gin.Context) {
	h.handleSubscription(ctx, false)
}

func (h *SubscriptionHandler) handleSubscription(ctx *gin.Context, subscribe bool) {
	userID := common.GetUserIDFromContextPayload(ctx)
	eventID, err := h.idParser.ParseEventID(ctx)
	if err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	token := ctx.Query("invite")

	params := models.SubscriptionParams{
		EventID: eventID,
		UserID:  userID,
		Token:   token,
	}

	var Event models.Event
	if subscribe {
		Event, err = h.service.Subscribe(ctx, params)
		if err != nil {
			ctx.Error(apperror.BadRequest.WithCause(err))
			return
		}
	} else {
		Event, err = h.service.Unsubscribe(ctx, params)
	}

	if err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}

	resp := h.converter.ToEventResponse(Event)

	ctx.JSON(http.StatusOK, resp)
}
