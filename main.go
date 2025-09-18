package main

import (
	"clipboard-server/auth"
	"clipboard-server/config"
	"clipboard-server/database"
	"clipboard-server/handlers"
	"clipboard-server/middleware"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

var startTime = time.Now()

func main() {
	cfg := config.LoadConfig()

	if err := cfg.Validate(); err != nil {
		log.Fatal("Configuration validation failed:", err)
	}

	cfg.Print()

	if err := database.Initialize(); err != nil {
		log.Fatal("Database initialization failed:", err)
	}
	defer database.Close()

	if err := database.CreateIndexes(); err != nil {
		log.Printf("Failed to create database indexes: %v", err)
	}

	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	router := gin.New()

	setupMiddleware(router)
	setupRoutes(router)

	server := &http.Server{
		Addr:           cfg.GetAddress(),
		Handler:        router,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	go func() {
		fmt.Printf("Server starting on http://%s\n", cfg.GetAddress())
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed to start:", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	} else {
		fmt.Println("Server gracefully stopped")
	}
}

func setupMiddleware(router *gin.Engine) {
	router.Use(middleware.HealthCheck())
	router.Use(middleware.RequestID())
	router.Use(middleware.ErrorHandler())
	router.Use(middleware.Security())
	router.Use(middleware.SetupCORS())
	router.Use(middleware.RateLimit())
	router.Use(middleware.ContentSizeLimit())
	router.Use(middleware.RequestLogger())
}

func setupRoutes(router *gin.Engine) {
	v1 := router.Group("/api/v1")

	authHandler := handlers.NewAuthHandler()
	clipboardHandler := handlers.NewClipboardHandler()

	authGroup := v1.Group("/auth")
	{
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/refresh", authHandler.RefreshToken)
	}

	authenticatedGroup := v1.Group("/")
	authenticatedGroup.Use(auth.JWTAuthMiddleware())
	{
		userGroup := authenticatedGroup.Group("/user")
		{
			userGroup.GET("/profile", authHandler.GetProfile)
			userGroup.POST("/logout", authHandler.Logout)
		}

		clipboardGroup := authenticatedGroup.Group("/clipboard")
		{
			clipboardGroup.GET("/items", clipboardHandler.GetItems)
			clipboardGroup.POST("/items", clipboardHandler.CreateItem)
			clipboardGroup.GET("/items/:id", clipboardHandler.GetItem)
			clipboardGroup.PUT("/items/:id", clipboardHandler.UpdateItem)
			clipboardGroup.DELETE("/items/:id", clipboardHandler.DeleteItem)
			clipboardGroup.POST("/sync", clipboardHandler.BatchSync)
			clipboardGroup.POST("/sync-single", clipboardHandler.SyncSingleItem) // 新增单项同步接口
			clipboardGroup.GET("/statistics", clipboardHandler.GetStatistics)
			clipboardGroup.GET("/recent", clipboardHandler.GetRecentSyncItems) // 新增最近同步接口
			clipboardGroup.GET("/latest", clipboardHandler.GetLatestSyncItem) // 新增获取最新单条记录接口
		}
	}

	systemGroup := v1.Group("/system")
	{
		systemGroup.GET("/health", healthCheck)
		systemGroup.GET("/info", systemInfo)
		systemGroup.GET("/stats", systemStats)
	}

	router.GET("/", rootHandler)
	router.NoRoute(notFoundHandler)
}

func rootHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"service":     "clipboard-sync-server",
		"version":     "1.0.0",
		"status":      "running",
		"timestamp":   time.Now().Format(time.RFC3339),
		"api_version": "v1",
		"endpoints": gin.H{
			"auth":      "/api/v1/auth",
			"clipboard": "/api/v1/clipboard",
			"system":    "/api/v1/system",
			"health":    "/api/v1/system/health",
		},
	})
}

func healthCheck(c *gin.Context) {
	dbStatus := "ok"
	if err := database.HealthCheck(); err != nil {
		dbStatus = "error: " + err.Error()
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().Format(time.RFC3339),
		"service":   "clipboard-sync-server",
		"version":   "1.0.0",
		"database":  dbStatus,
		"uptime":    time.Since(startTime).String(),
	})
}

func systemInfo(c *gin.Context) {
	cfg := config.GetConfig()

	c.JSON(http.StatusOK, gin.H{
		"service":     "clipboard-sync-server",
		"version":     "1.0.0",
		"environment": os.Getenv("GO_ENV"),
		"config": gin.H{
			"max_content_size": cfg.MaxContentSize,
			"cleanup_days":     cfg.CleanupDays,
			"rate_limit_rps":   cfg.RateLimitRPS,
			"rate_limit_burst": cfg.RateLimitBurst,
			"upload_max_size":  cfg.UploadMaxSize,
		},
		"timestamp": time.Now().Format(time.RFC3339),
		"uptime":    time.Since(startTime).String(),
	})
}

func systemStats(c *gin.Context) {
	dbStats := database.GetStats()

	c.JSON(http.StatusOK, gin.H{
		"timestamp": time.Now().Format(time.RFC3339),
		"uptime":    time.Since(startTime).String(),
		"database":  dbStats,
	})
}

func notFoundHandler(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{
		"error":   "not found",
		"message": "the requested resource was not found",
		"path":    c.Request.URL.Path,
	})
}
