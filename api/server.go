package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"treffly/api/handler"
	eventservice "treffly/api/service/event"
	geoservice "treffly/api/service/geo"
	tagservice "treffly/api/service/tag"
	tokenservice "treffly/api/service/token"
	userservice "treffly/api/service/user"
	db "treffly/db/sqlc"
	"treffly/token"
	"treffly/util"
)

type Server struct {
	config         util.Config
	store          db.Store
	tokenMaker     token.Maker
	router         *gin.Engine
	geocodeClient *geoservice.GeocoderClient
	suggestClient  *geoservice.SuggestClient
}

func NewServer(config util.Config, store db.Store) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	geocoderClient := geoservice.NewGeocoderClient(config.YandexGeocoderAPIKey)
	suggesterClient := geoservice.NewSuggestClient(config.YandexSuggesterAPIKey)
	server := &Server{
		store:      store,
		tokenMaker: tokenMaker,
		config:     config,
		geocodeClient: geocoderClient,
		suggestClient: suggesterClient,
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

	eventService := eventservice.New(server.store)
	eventHandler := handler.NewEventHandler(eventService)

	userService := userservice.New(server.store, server.tokenMaker, server.config)
	userHandler := handler.NewUserHandler(userService, server.config)

	tagService := tagservice.New(server.store)
	tagHandler := handler.NewTagHandler(tagService)

	geoService := geoservice.New(server.store, server.geocodeClient, server.suggestClient)
	geoHandler := handler.NewGeoHandler(geoService)

	tokenService := tokenservice.New(server.store, server.tokenMaker, server.config)
	tokenHandler := handler.NewTokenHandler(tokenService, server.config)

	router.POST("/users", userHandler.Create)
	router.POST("/login", userHandler.Login)
	router.POST("/auth/refresh", tokenHandler.RefreshTokens)
	router.GET("/auth", userHandler.Auth)
	router.GET("/tags", tagHandler.GetTags)
	router.GET("/events", eventHandler.List)

	router.GET("/geocode", geoHandler.Geocode)
	router.GET("/suggest/addresses", geoHandler.Suggest)
	router.GET("/reverse-geocode", geoHandler.ReverseGeocode)

	softAuthRoutes := router.Group("/").Use(softAuthMiddleware(server.tokenMaker))
	softAuthRoutes.GET("/events/home", eventHandler.GetHome)
	softAuthRoutes.GET("/events/:id", eventHandler.GetByID)

	authRoutes := router.Group("/").Use(authMiddleware(server.tokenMaker))
	authRoutes.POST("/logout", userHandler.Logout)

	authRoutes.GET("/users/me", userHandler.GetCurrent)
	authRoutes.PUT("/users/me", userHandler.UpdateCurrent)
	authRoutes.DELETE("/users/me", userHandler.DeleteCurrent)
	authRoutes.PUT("users/me/tags", userHandler.UpdateCurrentTags)

	authRoutes.POST("/events", eventHandler.Create)
	authRoutes.PUT("/events/:id", eventHandler.Update)
	authRoutes.DELETE("/events/:id", eventHandler.Delete)
	authRoutes.POST("/events/:id/subscription", eventHandler.Subscribe)
	authRoutes.DELETE("/events/:id/subscription", eventHandler.Unsubscribe)
	authRoutes.GET("/users/me/past-events", eventHandler.GetPast)
	authRoutes.GET("/users/me/upcoming-events", eventHandler.GetUpcoming)
	authRoutes.GET("/users/me/owned-events", eventHandler.GetOwned)

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
		err := v.RegisterValidation("valid_date", validDate)
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
