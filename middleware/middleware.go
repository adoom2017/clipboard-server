package middleware

import (
	"clipboard-server/config"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/time/rate"
)

// SetupCORS configures CORS middleware
func SetupCORS() gin.HandlerFunc {
	cfg := config.GetConfig()

	corsConfig := cors.Config{
		AllowOrigins:     cfg.CORSAllowOrigins,
		AllowMethods:     cfg.CORSAllowMethods,
		AllowHeaders:     cfg.CORSAllowHeaders,
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	// If all origins allowed, use AllowAllOrigins
	if len(cfg.CORSAllowOrigins) == 1 && cfg.CORSAllowOrigins[0] == "*" {
		corsConfig.AllowAllOrigins = true
		corsConfig.AllowOrigins = nil
	}

	return cors.New(corsConfig)
}

// RateLimit middleware for rate limiting
func RateLimit() gin.HandlerFunc {
	cfg := config.GetConfig()
	limiter := rate.NewLimiter(rate.Limit(cfg.RateLimitRPS), cfg.RateLimitBurst)

	return gin.HandlerFunc(func(c *gin.Context) {
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "rate limit exceeded",
				"message": "too many requests, please slow down",
			})
			c.Abort()
			return
		}
		c.Next()
	})
}

// RequestLogger middleware for request logging
func RequestLogger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("[%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.TimeStamp.Format("2006/01/02 15:04:05"),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ClientIP,
		)
	})
}

// ErrorHandler middleware for error handling
func ErrorHandler() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal server error",
				"message": err,
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal server error",
				"message": "an unexpected error occurred",
			})
		}
		c.Abort()
	})
}

// Security middleware for security headers
func Security() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Next()
	}
}

// ContentSizeLimit middleware for content size limiting
func ContentSizeLimit() gin.HandlerFunc {
	cfg := config.GetConfig()
	maxSize := cfg.MaxContentSize

	return func(c *gin.Context) {
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			if c.Request.ContentLength > maxSize {
				c.JSON(http.StatusRequestEntityTooLarge, gin.H{
					"error":   "request entity too large",
					"message": fmt.Sprintf("request body size exceeds limit of %d bytes", maxSize),
				})
				c.Abort()
				return
			}
		}
		c.Next()
	}
}

// APIKeyAuth middleware for API key authentication (optional)
func APIKeyAuth() gin.HandlerFunc {
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		// Skip if no API key set
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		key := c.GetHeader("X-API-Key")
		if key == "" {
			key = c.Query("api_key")
		}

		if key != apiKey {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "invalid or missing API key",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequestID middleware to generate unique request IDs
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		c.Header("X-Request-ID", requestID)
		c.Set("RequestID", requestID)
		c.Next()
	}
}

// generateRequestID generates a request ID
func generateRequestID() string {
	// Simple timestamp + random UUID approach
	return fmt.Sprintf("%d-%s",
		time.Now().UnixNano(),
		strings.ReplaceAll(uuid.New().String()[:8], "-", ""),
	)
}

// HealthCheck middleware for health checks
func HealthCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/health" {
			c.JSON(http.StatusOK, gin.H{
				"status":    "ok",
				"timestamp": time.Now().Format(time.RFC3339),
				"service":   "clipboard-sync-server",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
