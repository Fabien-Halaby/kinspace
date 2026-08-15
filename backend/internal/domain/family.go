package domain

import (
	"context"
	"time"
)

// Family is the tenant that scopes every domain object. All access to
// data within a family is isolated server-side through family_id.
type Family struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	InviteCode string    `json:"invite_code"`
	CreatedAt  time.Time `json:"created_at"`
}

// FamilyRepository is the storage port for the family aggregate.
type FamilyRepository interface {
	Create(ctx context.Context, name, inviteCode string, ownerID int64) (Family, error)
	FindByUserID(ctx context.Context, userID int64) (Family, error)
	Join(ctx context.Context, userID int64, inviteCode string) (Family, error)
}
