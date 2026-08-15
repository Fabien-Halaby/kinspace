package httpapi

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/Fabien-Halaby/kinspace/backend/internal/domain"
	"github.com/Fabien-Halaby/kinspace/backend/internal/token"
	"github.com/gin-gonic/gin"
)

// Error writes a JSON error response whose HTTP status is derived from
// the domain error. It is the single place where errors are mapped to
// status codes, keeping handlers free of status-code logic.
func Error(c *gin.Context, err error) {
	var validation domain.ValidationError

	switch {
	case errors.As(err, &validation):
		writeError(c, http.StatusBadRequest, validation.Error())
	case errors.Is(err, domain.ErrEmailExists):
		writeError(c, http.StatusConflict, domain.ErrEmailExists.Error())
	case errors.Is(err, domain.ErrInvalidCredentials):
		writeError(c, http.StatusUnauthorized, "invalid credentials")
	case errors.Is(err, domain.ErrUnauthorized):
		writeError(c, http.StatusUnauthorized, domain.ErrUnauthorized.Error())
	case errors.Is(err, domain.ErrAlreadyInFamily):
		writeError(c, http.StatusConflict, domain.ErrAlreadyInFamily.Error())
	case errors.Is(err, domain.ErrFamilyNotFound):
		writeError(c, http.StatusNotFound, domain.ErrFamilyNotFound.Error())
	case errors.Is(err, domain.ErrInviteCodeInvalid):
		writeError(c, http.StatusNotFound, domain.ErrInviteCodeInvalid.Error())
	case errors.Is(err, domain.ErrUserNotFound):
		writeError(c, http.StatusNotFound, domain.ErrUserNotFound.Error())
	case errors.Is(err, domain.ErrNotInSameFamily):
		writeError(c, http.StatusForbidden, domain.ErrNotInSameFamily.Error())
	case errors.Is(err, domain.ErrInvalidRelationType), errors.Is(err, domain.ErrSelfRelation):
		writeError(c, http.StatusBadRequest, err.Error())
	case errors.Is(err, token.ErrExpiredToken):
		writeError(c, http.StatusUnauthorized, token.ErrExpiredToken.Error())
	case errors.Is(err, token.ErrInvalidToken):
		writeError(c, http.StatusUnauthorized, token.ErrInvalidToken.Error())
	default:
		// Log the real cause server-side without leaking internals to
		// the client.
		slog.Error("unhandled error", "error", err)
		writeError(c, http.StatusInternalServerError, "internal server error")
	}
}

// BadRequest writes a plain 400 response, used for request-shape
// failures (malformed JSON, missing fields).
func BadRequest(c *gin.Context, message string) {
	writeError(c, http.StatusBadRequest, message)
}

func writeError(c *gin.Context, status int, message string) {
	c.AbortWithStatusJSON(status, gin.H{"error": message})
}
