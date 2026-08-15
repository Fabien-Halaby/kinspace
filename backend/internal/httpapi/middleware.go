package httpapi

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"strings"
	"time"

	"github.com/Fabien-Halaby/kinspace/backend/internal/domain"
	"github.com/Fabien-Halaby/kinspace/backend/internal/token"
	"github.com/gin-gonic/gin"
)

// RequestLogger logs every request with a correlation id that is also
// returned to the client so support issues can be traced.
func RequestLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		requestID := newRequestID()
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		c.Next()

		logger.Info("request",
			"request_id", requestID,
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"duration_ms", time.Since(start).Milliseconds(),
			"client_ip", c.ClientIP(),
		)
	}
}

// SecurityHeaders sets baseline hardening headers on every response.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "no-referrer")
		c.Header("Cache-Control", "no-store")
		c.Next()
	}
}

// RequireAuth authenticates the bearer token and injects the user
// identifier into the request context.
func RequireAuth(tokens token.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		parts := strings.Fields(header)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			Error(c, domain.ErrUnauthorized)
			c.Abort()
			return
		}
		userID, err := tokens.Parse(parts[1])
		if err != nil {
			Error(c, err)
			c.Abort()
			return
		}
		WithUserID(c, userID)
		c.Next()
	}
}

func newRequestID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return time.Now().Format("20060102150405.000000000")
	}
	return hex.EncodeToString(buf)
}
