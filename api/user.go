package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
	db "treffly/db/sqlc"
	"treffly/util"
)

type CreateUserRequest struct {
	Username string `json:"username" binding:"required,alphanum,min=2,max=16"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6,max=32"`
}

type userResponse struct {
	Username          string    `json:"username"`
	Email             string    `json:"email"`
	CreatedAt         time.Time `json:"created_at"`
}

func newUserResponse(user db.User) userResponse {
	return userResponse{
		Username:          user.Username,
		Email:             user.Email,
		CreatedAt:         user.CreatedAt,
	}
}

func (server *Server) createUser(ctx *gin.Context) {
	var req CreateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	arg := db.CreateUserParams{
		Username: req.Username,
		Email:    req.Email,
		PasswordHash: hashedPassword,
	}

	user, err := server.store.CreateUser(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	resp := newUserResponse(user)

	ctx.JSON(http.StatusOK, resp)
}
