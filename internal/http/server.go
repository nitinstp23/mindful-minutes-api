package http

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hellofresh/health-go/v5"
	"github.com/joho/godotenv"

	"github.com/mindful-minutes/mindful-minutes-api/internal/database"
)

type Server struct {
	router *gin.Engine
	port   string
}

func NewServer() *Server {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}

	// Set Gin mode
	ginMode := os.Getenv("GIN_MODE")
	if ginMode != "" {
		gin.SetMode(ginMode)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &Server{
		router: gin.Default(),
		port:   port,
	}

	// Connect to database
	if err := database.Connect(); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	server.setupHealthChecks()
	server.setupRoutes()
	return server
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
	// API routes
	api := s.router.Group("/api")
	{
		api.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "pong",
			})
		})
	}
}

func (s *Server) Start() error {
	log.Printf("Server starting on port %s", s.port)
	return s.router.Run(":" + s.port)
}
