package main

import (
	"log"
	"scandata/config"
	"scandata/database"
	"scandata/handlers"
	"scandata/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load config from .env
	cfg := config.LoadConfig()

	// Set Gin mode
	gin.SetMode(cfg.GinMode)

	// Initialize database
	database.InitDB(cfg)

	// Initialize security logger
	middleware.InitSecurityLogger(cfg)

	// Setup Gin
	r := gin.New()

	// Recovery middleware
	r.Use(gin.Recovery())

	// Security middlewares (OWASP)
	r.Use(middleware.ErrorHandlerMiddleware(cfg.IsProduction()))
	r.Use(middleware.SecurityHeadersMiddleware())
	r.Use(middleware.RequestSizeLimitMiddleware(cfg.MaxBodySize))
	r.Use(middleware.RateLimitMiddleware(cfg))
	r.Use(middleware.SecurityLoggerMiddleware(cfg))

	// CORS configuration (A01: Broken Access Control fix)
	corsConfig := cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}

	// In debug mode, allow all origins for easier development
	if cfg.IsProduction() {
		corsConfig.AllowOrigins = cfg.AllowedOrigins
	} else {
		corsConfig.AllowAllOrigins = true
	}
	r.Use(cors.New(corsConfig))

	// Custom logger
	r.Use(gin.Logger())

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(cfg)
	userHandler := handlers.NewUserHandler()
	unitHandler := handlers.NewUnitHandler()
	scanHandler := handlers.NewScanHandler()
	reportHandler := handlers.NewReportHandler()

	// Public routes
	r.POST("/api/auth/login", authHandler.Login)

	// Protected routes
	protected := r.Group("/api")
	protected.Use(middleware.AuthMiddleware(cfg))
	{
		// Auth
		protected.GET("/auth/me", authHandler.Me)
		protected.POST("/auth/change-password", authHandler.ChangePassword)

		// Users (Admin only)
		adminRoutes := protected.Group("/users")
		adminRoutes.Use(middleware.AdminMiddleware())
		{
			adminRoutes.GET("", userHandler.List)
			adminRoutes.POST("", userHandler.Create)
			adminRoutes.GET("/:id", userHandler.Get)
			adminRoutes.PUT("/:id", userHandler.Update)
			adminRoutes.DELETE("/:id", userHandler.Delete)
		}

		// Units (Admin only for CUD, all for Read)
		protected.GET("/units", unitHandler.List)
		protected.GET("/units/:id", unitHandler.Get)
		protected.GET("/units/qr/:qr_code", unitHandler.GetByQRCode)

		unitAdminRoutes := protected.Group("/units")
		unitAdminRoutes.Use(middleware.AdminMiddleware())
		{
			unitAdminRoutes.POST("", unitHandler.Create)
			unitAdminRoutes.PUT("/:id", unitHandler.Update)
			unitAdminRoutes.DELETE("/:id", unitHandler.Delete)
		}

		// Scans
		protected.POST("/scans", scanHandler.Submit)
		protected.GET("/scans", scanHandler.List)
		protected.GET("/scans/stats", scanHandler.GetStats)

		// Reports
		protected.GET("/reports/summary", reportHandler.Summary)
		protected.GET("/reports/daily", reportHandler.Daily)

		reportAdminRoutes := protected.Group("/reports")
		reportAdminRoutes.Use(middleware.AdminMiddleware())
		{
			reportAdminRoutes.GET("/users", reportHandler.UserPerformance)
			reportAdminRoutes.GET("/export", reportHandler.Export)
		}
	}

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"mode":   cfg.GinMode,
		})
	})

	log.Printf("Server running on port %s (mode: %s)", cfg.ServerPort, cfg.GinMode)
	r.Run(":" + cfg.ServerPort)
}
