package main

import (
	"context"
	"log"
	"fmt"

	"github.com/gin-gonic/gin"
	//jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/devops-360-online/go-with-me/config"
	"github.com/devops-360-online/go-with-me/internal/handlers"
	"github.com/devops-360-online/go-with-me/internal/middlewares"
	"github.com/devops-360-online/go-with-me/internal/models"
	"github.com/devops-360-online/go-with-me/internal/repositories"
	"github.com/devops-360-online/go-with-me/internal/tracing"
	"github.com/devops-360-online/go-with-me/internal/websockets"
	"github.com/devops-360-online/go-with-me/internal/logger"
	//"gorm.io/gorm"
)

func main() {
	cfg := config.LoadConfig()
	router := gin.Default()

	// Apply middlewares
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Store cfg in context
	router.Use(func(c *gin.Context) {
		c.Set("config", cfg)
		c.Next()
	})

	// Initialize the logger
    if err := logger.InitLogger(); err != nil {
        log.Fatalf("Could not initialize logger: %v", err)
    }
    defer logger.CloseLogger()

	// Initialize the tracer
	tp, err := tracing.InitTracer()
	if err != nil {
        logger.LogMessage("fatal", fmt.Sprintf("Failed to initialize tracer: %v", err), "", nil)
	}
	// Ensure the tracer provider is shutdown gracefully
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			logger.LogMessage("fatal", fmt.Sprintf("Failed to shutdown tracer provider: %v", err), "", nil)

		}
	}()

	// Apply Tracing Middleware
	router.Use(middlewares.TracingMiddleware())

	// Initialize database
	db, err := repositories.NewDatabase(cfg)
	if err != nil {
		logger.LogMessage("fatal", fmt.Sprintf("Failed to connect to database: %v", err), "", nil)
	}

	// Migrate the schema
	db.AutoMigrate(&models.User{}, &models.Event{})

	// Add db to context
	router.Use(func(c *gin.Context) {
		c.Set("db", db)
		c.Next()
	})

	// Initialize authentication middleware
	authMiddleware, err := middlewares.AuthMiddleware()
	if err != nil {
		logger.LogMessage("fatal", fmt.Sprintf("JWT Error: %v", err), "", nil)
	}

	// Public routes
	router.POST("/register", handlers.RegisterHandler)
	router.POST("/login", authMiddleware.LoginHandler)

	// Protected routes
	auth := router.Group("/")
	auth.Use(authMiddleware.MiddlewareFunc())
	{
		auth.POST("/events", handlers.CreateEventHandler)
		auth.GET("/events", handlers.ListEventsHandler)
		auth.GET("/events/:id", handlers.GetEventHandler)
		auth.POST("/events/:id/join", handlers.JoinEventHandler)
		auth.DELETE("/events/:id/join", handlers.UnjoinEventHandler)
		auth.GET("/ws/events/:id", func(c *gin.Context) {
			handlers.EventChatHandler(c)
		})
	}

	// Initialize Redis client
	rdb := repositories.NewRedisClient(cfg)
	router.Use(middlewares.CacheMiddleware(rdb))

	// Initialize MongoDB client
	mongoClient, err := repositories.NewMongoClient(cfg)
	if err != nil {
		logger.LogMessage("fatal", fmt.Sprintf("Failed to connect to MongoDB: %v", err), "", nil)
	}

	// Add MongoDB client to context
	router.Use(func(c *gin.Context) {
		c.Set("mongoClient", mongoClient)
		c.Next()
	})

	// Start WebSocket hub
	go websockets.RunHub(rdb)

	// Start the server
	if err := router.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Failed to run server: %v", err)
		logger.LogMessage("fatal", fmt.Sprintf("Failed to run server: %v", err), "", nil)
	}
}
