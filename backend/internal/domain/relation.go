package domain

import (
	"context"
	"strings"
	"time"
)

// Relation types supported by the family graph. Each edge is stored once
// and its inverse (parent <-> child) is derived server-side.
const (
	RelationParent  = "parent"
	RelationChild   = "child"
	RelationSpouse  = "spouse"
	RelationSibling = "sibling"
)

// IsValidRelationType reports whether t is a supported relation type.
func IsValidRelationType(t string) bool {
	switch strings.ToLower(strings.TrimSpace(t)) {
	case RelationParent, RelationChild, RelationSpouse, RelationSibling:
		return true
	default:
		return false
	}
}

// Relation is a typed, directed edge in the family graph.
type Relation struct {
	ID            int64     `json:"id"`
	FamilyID      int64     `json:"family_id"`
	UserID        int64     `json:"user_id"`
	RelatedUserID int64     `json:"related_user_id"`
	Type          string    `json:"type"`
	CreatedAt     time.Time `json:"created_at"`
}

// RelationRepository is the storage port for the family graph.
type RelationRepository interface {
	Create(ctx context.Context, familyID, userID, relatedUserID int64, relationType string) (Relation, error)
	List(ctx context.Context, familyID int64) ([]Relation, error)
	UserFamilyID(ctx context.Context, userID int64) (int64, error)
}
