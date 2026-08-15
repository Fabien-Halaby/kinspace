package httpapi

import (
	"net/http"

	"github.com/Fabien-Halaby/kinspace/backend/internal/domain"
	"github.com/gin-gonic/gin"
)

// AuthHandler exposes the auth use cases over HTTP. It only talks to the
// AuthService port, never to storage or the token implementation.
type AuthHandler struct {
	service AuthService
}

type registerRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request body")
		return
	}

	user, signedToken, err := h.service.Register(c.Request.Context(), req.Name, req.Email, req.Password)
	if err != nil {
		Error(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"user": user, "token": signedToken})
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request body")
		return
	}

	user, signedToken, err := h.service.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		Error(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user, "token": signedToken})
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID, ok := UserID(c)
	if !ok {
		Error(c, domain.ErrUnauthorized)
		return
	}

	user, err := h.service.Me(c.Request.Context(), userID)
	if err != nil {
		Error(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user})
}
