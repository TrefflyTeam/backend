package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	eventdto "treffly/api/dto/event"
	userdto "treffly/api/dto/user"
	"treffly/api/handler/event"
	"treffly/api/handler/geo"
	image2 "treffly/api/handler/image"
	"treffly/api/handler/tag"
	token2 "treffly/api/handler/token"
	"treffly/api/handler/user"
	eventservice "treffly/api/service/event"
	"treffly/api/service/generator"
	geoservice "treffly/api/service/geo"
	imageservice "treffly/api/service/image"
	"treffly/api/service/mail"
	tagservice "treffly/api/service/tag"
	tokenservice "treffly/api/service/token"
	userservice "treffly/api/service/user"
	"treffly/db/redis"
	db "treffly/db/sqlc"
	"treffly/image"
	"treffly/logger"
	"treffly/token"
	"treffly/util"
)

type Server struct {
	config        util.Config
	store         db.Store
	tokenMaker    token.Maker
	router        *gin.Engine
	geocodeClient *geoservice.GeocoderClient
	suggestClient *geoservice.SuggestClient
	imageStore    image.Store
	rlClient      *redis.Client
}

func NewServer(config util.Config, store db.Store) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	geocoderClient := geoservice.NewGeocoderClient(config.YandexGeocoderAPIKey)
	suggesterClient := geoservice.NewSuggestClient(config.YandexSuggesterAPIKey)

	imageStore, err := image.NewLocalStorage(config.ImageBasePath)
	if err != nil {
		return nil, fmt.Errorf("cannot create image store: %w", err)
	}

	rlClient, err := redis.NewClient(&redis.Config{
		Host:     config.RedisHost,
		Port:     config.RedisPort,
		Password: config.RedisPassword,
		DB:       config.RedisDB,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot create redis store: %w", err)
	}

	server := &Server{
		store:         store,
		tokenMaker:    tokenMaker,
		config:        config,
		geocodeClient: geocoderClient,
		suggestClient: suggesterClient,
		imageStore:    imageStore,
		rlClient:      rlClient,
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

	log := logger.NewZapLogger(server.config.Environment)

	router.Use(ErrorHandler(log))

	eventConverter := eventdto.NewEventConverter(server.config.Environment, server.config.Domain)
	userConverter := userdto.NewUserConverter(server.config.Environment, server.config.Domain)

	imageService := imageservice.New(server.imageStore, server.config, server.store)

	generatorClient := generator.NewClient(server.config.GenBaseURL, server.config.GenAPIKey, server.config.GenSystemPrompt, server.config.GenModel)
	generatorHandler := event.NewGenerator(generatorClient)

	eventService := eventservice.New(server.store, server.config)
	eventQueryHandler := event.NewEventQueryHandler(eventService, imageService, eventConverter)
	eventCRUDHandler := event.NewEventCRUDHandler(eventService, imageService, eventConverter)
	eventSubscriptionHandler := event.NewEventSubscriptionHandler(eventService, eventConverter)

	resetStore := redis.NewRedisResetStore(server.rlClient)
	rlStore := redis.NewRateLimitStore(server.rlClient)

	userService := userservice.New(server.store, resetStore,  server.tokenMaker, server.config, rlStore)
	userProfileHandler := user.NewProfileHandler(userService, userService, userService, imageService, userConverter, server.config.Environment)
	userAuthHandler := user.NewAuthHandler(userService, userService, userConverter, server.config)

	tagService := tagservice.New(server.store)
	tagHandler := tag.NewTagHandler(tagService)

	geoService := geoservice.New(server.store, server.geocodeClient, server.suggestClient)
	geoHandler := geo.NewGeoHandler(geoService)

	tokenService := tokenservice.New(server.store, server.tokenMaker, server.config, log)
	tokenHandler := token2.NewTokenHandler(tokenService, server.config)

	imageHandler := image2.NewImageHandler(imageService)

	router.POST("/users", userAuthHandler.Create)
	router.POST("/login", userAuthHandler.Login)
	router.POST("/auth/refresh", tokenHandler.RefreshTokens)
	router.GET("/auth", tokenHandler.Auth)
	router.GET("/tags", tagHandler.GetTags)
	router.GET("/events", eventCRUDHandler.List)

	router.GET("/images/*path", imageHandler.Get)

	router.GET("/geocode", geoHandler.Geocode)
	router.GET("/suggest/addresses", geoHandler.Suggest)
	router.GET("/reverse-geocode", geoHandler.ReverseGeocode)

	softAuthRoutes := router.Group("/").Use(softAuthMiddleware(server.tokenMaker))
	softAuthRoutes.GET("/events/home", eventQueryHandler.GetHome)
	softAuthRoutes.GET("/events/:id", eventCRUDHandler.GetByID)

	authRoutes := router.Group("/").Use(authMiddleware(server.tokenMaker))
	authRoutes.POST("/logout", userAuthHandler.Logout)

	authRoutes.GET("/users/me", userProfileHandler.GetCurrent)
	authRoutes.PUT("/users/me", userProfileHandler.UpdateCurrent)
	authRoutes.DELETE("/users/me", userProfileHandler.DeleteCurrent)
	authRoutes.PUT("users/me/tags", userProfileHandler.UpdateCurrentTags)

	authRoutes.POST("/events", eventCRUDHandler.Create)
	authRoutes.PUT("/events/:id", eventCRUDHandler.Update)
	authRoutes.DELETE("/events/:id", eventCRUDHandler.Delete)
	authRoutes.POST("/events/:id/subscription", eventSubscriptionHandler.Subscribe)
	authRoutes.DELETE("/events/:id/subscription", eventSubscriptionHandler.Unsubscribe)
	authRoutes.GET("/users/me/past-events", eventQueryHandler.GetPast)
	authRoutes.GET("/users/me/upcoming-events", eventQueryHandler.GetUpcoming)
	authRoutes.GET("/users/me/owned-events", eventQueryHandler.GetOwned)
	authRoutes.GET("/events/:id/invite", tokenHandler.CreatePrivateEventToken)

	limitCheckHandler := user.NewLimitCheckHandler(&rlStore, server.config.GenLimit, server.config.GenTimeout)

	authRoutes.GET("/events/generate-desc", RateLimitMiddleware(&rlStore, server.config.GenLimit, server.config.GenTimeout), generatorHandler.CreateChatCompletion)
	authRoutes.POST("/users/generate-limit", limitCheckHandler.CheckGenerateRateLimit)

	mailer := mail.New(mail.SMTPConfig{
		Host: server.config.SMTPHost,
		Port: server.config.SMTPPort,
		Username: server.config.SMTPUsername,
		Password: server.config.SMTPPassword,
		DefaultFrom: server.config.SMTPDefaultFrom,
	})
	pwResetHandler := user.NewPasswordResetHandler(userService, mailer, server.config)
	router.POST("/forgot-pw", pwResetHandler.InitiatePasswordReset)
	router.POST("/verify-code", pwResetHandler.ConfirmResetCode)
	router.POST("/reset-pw", pwResetHandler.CompletePasswordReset)

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
