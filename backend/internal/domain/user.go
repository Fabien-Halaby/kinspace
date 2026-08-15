package domain

import (
	"context"
	"strings"
	"time"
)

// User is the aggregate root of the auth and member-profile domains.
// PasswordHash must never be serialized to clients.
type User struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	FamilyID     *int64    `json:"family_id,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// Normalize canonicalizes the mutable string fields so the same value
// cannot be stored twice under different casings.
func (u *User) Normalize() {
	u.Name = strings.TrimSpace(u.Name)
	u.Email = strings.ToLower(strings.TrimSpace(u.Email))
}

// UserRepository is the storage port for users. Implementations live in
// the adapters layer; the domain and application layers only depend on
// this interface.
type UserRepository interface {
	Create(ctx context.Context, name, email, passwordHash string) (User, error)
	FindByEmail(ctx context.Context, email string) (User, error)
	FindByID(ctx context.Context, id int64) (User, error)
}
