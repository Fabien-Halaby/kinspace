package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var ErrEmailExists = errors.New("email already exists")

var ErrInvalidCredentials = errors.New("invalid credentials")

type User struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
	FamilyID     *int64 `json:"family_id,omitempty"`
}

type Repository interface {
	Create(ctx context.Context, name, email, passwordHash string) (User, error)
	FindByEmail(ctx context.Context, email string) (User, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service { return &Service{repo: repo} }

func (s *Service) Register(ctx context.Context, name, email, password string) (User, error) {
	name = strings.TrimSpace(name)
	email = strings.ToLower(strings.TrimSpace(email))
	if len(name) < 2 || len(name) > 100 {
		return User{}, fmt.Errorf("name must be between 2 and 100 characters")
	}
	if !strings.Contains(email, "@") || len(email) > 254 {
		return User{}, fmt.Errorf("invalid email")
	}
	if len(password) < 8 {
		return User{}, fmt.Errorf("password must contain at least 8 characters")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, fmt.Errorf("hash password: %w", err)
	}

	user, err := s.repo.Create(ctx, name, email, string(hash))
	if err != nil {
		return User{}, err
	}
	return user, nil
}
