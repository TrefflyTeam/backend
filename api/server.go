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

	err = server.registerValidators()
	if err != nil {
		return nil, fmt.Errorf("cannot register validators: %w", err)
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
	authRoutes.DELETE("/users/me", server.deleteCurrentUser)
	authRoutes.POST("/users/me/tags/:id", server.addCurrentUserTag)
	authRoutes.DELETE("/users/me/tags/:id", server.deleteCurrentUserTag)

	authRoutes.POST("/events", server.createEvent)

	server.router = router
}

func (server *Server) registerValidators() error {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		err := v.RegisterValidation("username", validUsername)
		if err != nil {
			return err
		}
	}
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		err := v.RegisterValidation("event_name", validEventName)
		if err != nil {
			return err
		}
	}
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		err := v.RegisterValidation("latitude", validLatitude)
		if err != nil {
			return err
		}
	}
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		err := v.RegisterValidation("longitude", validLongitude)
		if err != nil {
			return err
		}
	}
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		err := v.RegisterValidation("date", validDate)
		if err != nil {
			return err
		}
	}
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		err := v.RegisterValidation("positive", validPositiveInteger)
		if err != nil {
			return err
		}
	}
	return nil
}

func (server *Server) Start(address string) error {
	return server.router.Run(address)
}
