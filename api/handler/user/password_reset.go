package user

import (
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
	"treffly/api/common"
	userdto "treffly/api/dto/user"
	"treffly/apperror"
	"treffly/util"
)

type PasswordResetService interface {
	InitiatePasswordReset(ctx context.Context, email string) (string, error)
	ConfirmResetCode(ctx context.Context, email, code string) (string, error)
	CompletePasswordReset(ctx context.Context, token, newPassword string) error
}

type mailer interface {
	SendPasswordReset(to, resetCode string, expiry time.Duration) error
}

type PasswordResetHandler struct {
	service PasswordResetService
	mailer     mailer
	config util.Config
}

func NewPasswordResetHandler(service PasswordResetService, mailer mailer, config util.Config) *PasswordResetHandler {
	return &PasswordResetHandler{
		service: service,
		mailer:  mailer,
		config: config,
	}
}

func (h *PasswordResetHandler) InitiatePasswordReset(ctx *gin.Context) {
	var req userdto.InitiateResetRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	code, err := h.service.InitiatePasswordReset(ctx, req.Email)
	if err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	if code != "" {
		err = h.mailer.SendPasswordReset(req.Email, code, h.config.ResetCodeTTL)
	}

	ctx.JSON(http.StatusOK,  gin.H{"Message": "Код был успешно отправлен"})
}

func (h *PasswordResetHandler) ConfirmResetCode(ctx *gin.Context) {
	var req userdto.ConfirmResetRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	token, err := h.service.ConfirmResetCode(ctx, req.Email, req.Code)
	if err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	common.SetTokenCookie(ctx, "reset_token", token, common.ResetTokenCookiePath, h.config.ResetTokenDuration, h.config.Environment)

	ctx.JSON(http.StatusOK, gin.H{"Message": "Код подтверждён"})
}

func (h *PasswordResetHandler) CompletePasswordReset(ctx *gin.Context) {
	var req userdto.CompleteResetRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	token, err := ctx.Cookie("reset_token")
	if err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	if err := h.service.CompletePasswordReset(ctx, token, req.NewPassword); err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"Message": "Пароль успешно обновлён"})
}
