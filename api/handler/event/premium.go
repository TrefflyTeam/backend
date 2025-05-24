package event

import (
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"treffly/api/common"
	eventdto "treffly/api/dto/event"
	"treffly/api/models"
	"treffly/apperror"
)

type premiumService interface {
	CreatePremiumOrder(ctx context.Context, params models.PremiumOrderParams) (models.PremiumOrder, error)
	GetPremiumOrder(ctx context.Context, id int32) (models.PremiumOrder, error)
	CompletePremiumOrder(ctx context.Context, id int32) error
}

type PremiumHandler struct {
	service premiumService
	shop    string
	price   float64
}

func NewPremiumHandler(service premiumService, shop string, price float64) *PremiumHandler {
	return &PremiumHandler{
		service: service,
		shop:    shop,
		price:   price,
	}
}

func (h *PremiumHandler) CreatePremiumOrder(ctx *gin.Context) {
	userID := common.GetUserIDFromContextPayload(ctx)
	var req eventdto.CreatePremiumOrderRequest
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	arg := models.PremiumOrderParams{
		UserID:  userID,
		EventID: req.EventID,
		Shop:    h.shop,
		Price:   h.price,
	}

	order, err := h.service.CreatePremiumOrder(ctx, arg)
	if err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"id": order.ID})
}

func (h *PremiumHandler) GetPremiumOrder(ctx *gin.Context) {
	eventID := ctx.Param("id")
	id, err := strconv.Atoi(eventID)
	if err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	order, err := h.service.GetPremiumOrder(ctx, int32(id))
	if err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	resp := eventdto.PremiumOrderResponse{
		ID:        order.ID,
		EventID:   order.EventID,
		Shop:      order.Shop,
		Price:     order.Price,
		Status:    order.Status,
		CreatedAt: order.CreatedAt,
	}

	ctx.JSON(http.StatusOK, resp)
}

func (h *PremiumHandler) CompletePremiumOrder(ctx *gin.Context) {
	eventID := ctx.Param("id")
	id, err := strconv.Atoi(eventID)
	if err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	err = h.service.CompletePremiumOrder(ctx, int32(id))
	if err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "complete"})
}
