package httpapi

import (
	"github.com/gin-gonic/gin"
)

type contextKey string

const userIDKey contextKey = "authenticated_user_id"

// WithUserID stores the authenticated user identifier in the request
// context. It must only be called by the auth middleware.
func WithUserID(c *gin.Context, userID int64) {
	c.Set(string(userIDKey), userID)
}

// UserID returns the authenticated user identifier previously stored by
// WithUserID. The second return value is false when the request was not
// authenticated.
func UserID(c *gin.Context) (int64, bool) {
	value, ok := c.Get(string(userIDKey))
	if !ok {
		return 0, false
	}
	userID, ok := value.(int64)
	return userID, ok
}
