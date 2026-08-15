package httpapi

import (
	"net/http"

	"github.com/Fabien-Halaby/kinspace/backend/internal/domain"
	"github.com/gin-gonic/gin"
)

// RelationHandler exposes the family-graph use cases over HTTP. It needs
// the family service to resolve the caller's family before operating on
// relations, keeping every query tenant-scoped.
type RelationHandler struct {
	families  FamilyService
	relations RelationService
}

type createRelationRequest struct {
	RelatedUserID int64  `json:"related_user_id" binding:"required"`
	Type          string `json:"type" binding:"required"`
}

func (h *RelationHandler) Create(c *gin.Context) {
	userID, ok := UserID(c)
	if !ok {
		Error(c, domain.ErrUnauthorized)
		return
	}

	var req createRelationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request body")
		return
	}

	family, err := h.families.Me(c.Request.Context(), userID)
	if err != nil {
		Error(c, err)
		return
	}

	relation, err := h.relations.Create(c.Request.Context(), family.ID, userID, req.RelatedUserID, req.Type)
	if err != nil {
		Error(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"relation": relation})
}

func (h *RelationHandler) List(c *gin.Context) {
	userID, ok := UserID(c)
	if !ok {
		Error(c, domain.ErrUnauthorized)
		return
	}

	family, err := h.families.Me(c.Request.Context(), userID)
	if err != nil {
		Error(c, err)
		return
	}

	relations, err := h.relations.List(c.Request.Context(), family.ID)
	if err != nil {
		Error(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"relations": relations})
}
