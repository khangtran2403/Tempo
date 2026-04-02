package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/brokeboycoding/tempo/internal/config"
	"github.com/brokeboycoding/tempo/pkg/common"

	"github.com/brokeboycoding/tempo/internal/auth"

	"github.com/brokeboycoding/tempo/pkg/metrics"
	"github.com/brokeboycoding/tempo/pkg/ratelimit"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware middleware để xác minh JWT token
func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := common.GetLogger()
		var token string

		// First, try to get the token from the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				token = parts[1]
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "invalid authorization header format",
				})
				c.Abort()
				return
			}
		}

		// If token is not in the header, try to get it from query parameter
		// This is useful for redirect-based flows like OAuth connection initiation
		if token == "" {
			token = c.Query("token")
		}

		// If token is still empty, return error
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "missing authorization token",
			})
			c.Abort()
			return
		}


		// Xác minh token
		claims, err := auth.VerifyToken(token, cfg.JWT.Secret)
		if err != nil {
			logger.Debugf("Token verification failed: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or expired token",
			})
			c.Abort()
			return
		}

		// Lưu claims vào context
		// Các handler sau có thể lấy từ c.Get("user_id")
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)

		// Tiếp tục xử lý request
		c.Next()
	}
}
func RateLimitMiddleware(limiter *ratelimit.RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {

		identifier := getRateLimitIdentifier(c)

		limit, window := getRateLimitConfig(c.FullPath())

		allowed, remaining, resetTime, err := limiter.Allow(
			c.Request.Context(),
			identifier,
			limit,
			window,
		)

		if err != nil {

			logger := common.GetLogger()
			logger.Errorf("Rate limit check failed: %v", err)
			c.Next()
			return
		}

		c.Header("X-RateLimit-Limit", strconv.Itoa(limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))

		if !allowed {
			c.Header("Retry-After", strconv.FormatInt(int64(time.Until(resetTime).Seconds()), 10))

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate limit exceeded",
				"message":     "Too many requests. Please try again later.",
				"retry_after": time.Until(resetTime).Seconds(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// PrometheusMiddleware đo metrics cho tất cả HTTP requests
func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		metrics.HTTPRequestsInFlight.Inc()
		defer metrics.HTTPRequestsInFlight.Dec()

		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		c.Next()
		duration := time.Since(start).Seconds()
		method := c.Request.Method
		status := strconv.Itoa(c.Writer.Status())

		metrics.HTTPRequestsTotal.WithLabelValues(
			method,
			path,
			status,
		).Inc()

		metrics.HTTPRequestDuration.WithLabelValues(
			method,
			path,
		).Observe(duration)
	}
}

// Helper function để lấy user_id từ context
func GetUserID(c *gin.Context) string {
	userID, exists := c.Get("user_id")
	if !exists {
		return ""
	}
	return userID.(string)
}

// Helper function để lấy email từ context
func GetEmail(c *gin.Context) string {
	email, exists := c.Get("email")
	if !exists {
		return ""
	}
	return email.(string)
}
func getRateLimitIdentifier(c *gin.Context) string {
	if userID, exists := c.Get("user_id"); exists {
		return fmt.Sprintf("user:%s", userID)
	}

	return fmt.Sprintf("ip:%s", c.ClientIP())
}

// getRateLimitConfig trả về limit config cho từng endpoint
func getRateLimitConfig(path string) (int, time.Duration) {
	switch path {
	case "/api/v1/auth/register", "/api/v1/auth/login":

		return 5, 1 * time.Minute

	case "/api/v1/workflows/:id/trigger":

		return 60, 1 * time.Minute

	case "/api/v1/webhooks/:workflow_id":

		return 100, 1 * time.Minute

	default:

		return 100, 1 * time.Minute
	}
}
