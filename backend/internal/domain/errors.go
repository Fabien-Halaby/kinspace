package domain

import (
	"errors"
	"fmt"
)

// ValidationError is returned by the application layer when a business
// rule or input constraint is violated. It is mapped to HTTP 400.
type ValidationError struct {
	Field  string
	Reason string
}

func (e ValidationError) Error() string {
	if e.Field == "" {
		return e.Reason
	}
	return fmt.Sprintf("%s: %s", e.Field, e.Reason)
}

var (
	ErrEmailExists         = errors.New("email already registered")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrUserNotFound        = errors.New("user not found")
	ErrAlreadyInFamily     = errors.New("user already belongs to a family")
	ErrFamilyNotFound      = errors.New("family not found")
	ErrInviteCodeInvalid   = errors.New("invalid invite code")
	ErrInvalidRelationType = errors.New("invalid relation type")
	ErrNotInSameFamily     = errors.New("users must belong to the same family")
	ErrSelfRelation        = errors.New("cannot relate a user to themselves")
	ErrUnauthorized        = errors.New("unauthorized")
)
