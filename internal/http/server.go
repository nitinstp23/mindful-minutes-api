package http

import (
	"context"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hellofresh/health-go/v5"

	"github.com/mindful-minutes/mindful-minutes-api/internal/auth"
	"github.com/mindful-minutes/mindful-minutes-api/internal/config"
	"github.com/mindful-minutes/mindful-minutes-api/internal/database"
	"github.com/mindful-minutes/mindful-minutes-api/internal/handlers"
)

type Server struct {
	router *gin.Engine
	config *config.Config
}

func NewServer() (*Server, error) {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	// Set Gin mode
	gin.SetMode(cfg.Server.GinMode)

	server := &Server{
		router: gin.Default(),
		config: cfg,
	}

	// Connect to database
	if err := database.Connect(cfg.Database.URL); err != nil {
		return nil, err
	}

	server.setupHealthChecks()
	server.setupRoutes()

	return server, nil
}

func (s *Server) setupHealthChecks() {
	// Create health checker with database check
	h, _ := health.New(
		health.WithComponent(health.Component{
			Name:    "database",
			Version: "1.0.0",
		}),
		health.WithChecks(
			health.Config{
				Name:      "postgres",
				Timeout:   time.Second * 2,
				SkipOnErr: false,
				Check: func(ctx context.Context) error {
					return database.IsHealthy()
				},
			},
		),
	)

	// Liveness check - basic check that the service is running
	livenessChecker, _ := health.New()
	s.router.GET("/health/liveness", func(c *gin.Context) {
		livenessChecker.Handler().ServeHTTP(c.Writer, c.Request)
	})

	// Readiness check - check that the service is ready to serve traffic (includes DB)
	s.router.GET("/health/readiness", func(c *gin.Context) {
		h.Handler().ServeHTTP(c.Writer, c.Request)
	})
}

func (s *Server) setupRoutes() {
	// Webhooks (no auth required)
	webhooks := s.router.Group("/api/webhooks")
	{
		webhooks.POST("/clerk", auth.VerifyClerkWebhook(s.config))
	}

	// API routes
	api := s.router.Group("/api")
	{
		api.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "pong",
			})
		})
	}

	// Protected API routes (require authentication)
	protected := s.router.Group("/api")
	protected.Use(auth.AuthMiddleware(s.config))
	{
		// User routes
		protected.GET("/user/profile", handlers.GetUserProfile)
	}
}

func (s *Server) Start() error {
	log.Printf("Server starting on port %s", s.config.Server.Port)

	return s.router.Run(":" + s.config.Server.Port)
}
