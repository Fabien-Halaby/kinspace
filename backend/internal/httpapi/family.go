package httpapi

import (
	"net/http"

	"github.com/Fabien-Halaby/kinspace/backend/internal/domain"
	"github.com/gin-gonic/gin"
)

// FamilyHandler exposes the family use cases over HTTP.
type FamilyHandler struct {
	service FamilyService
}

type createFamilyRequest struct {
	Name string `json:"name" binding:"required"`
}

func (h *FamilyHandler) Create(c *gin.Context) {
	userID, ok := UserID(c)
	if !ok {
		Error(c, domain.ErrUnauthorized)
		return
	}

	var req createFamilyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request body")
		return
	}

	family, err := h.service.Create(c.Request.Context(), userID, req.Name)
	if err != nil {
		Error(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"family": family})
}

type joinFamilyRequest struct {
	InviteCode string `json:"invite_code" binding:"required"`
}

func (h *FamilyHandler) Join(c *gin.Context) {
	userID, ok := UserID(c)
	if !ok {
		Error(c, domain.ErrUnauthorized)
		return
	}

	var req joinFamilyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request body")
		return
	}

	family, err := h.service.Join(c.Request.Context(), userID, req.InviteCode)
	if err != nil {
		Error(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"family": family})
}

func (h *FamilyHandler) Me(c *gin.Context) {
	userID, ok := UserID(c)
	if !ok {
		Error(c, domain.ErrUnauthorized)
		return
	}

	family, err := h.service.Me(c.Request.Context(), userID)
	if err != nil {
		Error(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"family": family})
}
