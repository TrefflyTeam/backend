package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	db "treffly/db/sqlc"
	"treffly/token"
	"treffly/util"
)

type Server struct {
	config     util.Config
	store      db.Store
	tokenMaker token.Maker
	router     *gin.Engine
}

func NewServer(config util.Config, store db.Store) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}
	server := &Server{
		store:      store,
		tokenMaker: tokenMaker,
		config:     config,
	}

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		err := v.RegisterValidation("username", validUsername)
		if err != nil {
			return nil, err
		}
	}
	if server.config.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	server.setupRouter()
	return server, nil
}

func (server *Server) setupRouter() {
	router := gin.Default()

	router.Use(ErrorHandler())

	router.POST("/users", server.createUser)
	router.POST("/login", server.loginUser)
	router.POST("/auth/refresh", server.refreshTokens)
	router.GET("/auth", server.auth)
	router.GET("/tags", server.getTags)

	authRoutes := router.Group("/").Use(authMiddleware(server.tokenMaker))
	authRoutes.POST("/logout", server.logoutUser)
	authRoutes.GET("/users/me", server.getCurrentUser)
	authRoutes.PUT("/users/me", server.updateCurrentUser)

	server.router = router
}

func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

