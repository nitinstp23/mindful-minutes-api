package http

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/hellofresh/health-go/v5"
	"github.com/joho/godotenv"
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

	server.setupHealthChecks()
	server.setupRoutes()
	return server
}

func (s *Server) setupHealthChecks() {
	// Create health checker
	h, _ := health.New()

	// Liveness check - basic check that the service is running
	s.router.GET("/health/liveness", func(c *gin.Context) {
		h.Handler().ServeHTTP(c.Writer, c.Request)
	})

	// Readiness check - check that the service is ready to serve traffic
	// TODO: Add database connectivity check when database is configured
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