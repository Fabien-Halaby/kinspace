package application

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Fabien-Halaby/kinspace/backend/internal/domain"
	"github.com/Fabien-Halaby/kinspace/backend/internal/token"
	"golang.org/x/crypto/bcrypt"
)

const (
	minPasswordLength = 8
	minNameLength     = 2
	maxNameLength     = 100
	maxEmailLength    = 254
)

// AuthService implements the auth use cases. It depends only on the
// storage port (domain.UserRepository) and the token issuer port.
type AuthService struct {
	users  domain.UserRepository
	tokens token.Manager
}

func NewAuthService(users domain.UserRepository, tokens token.Manager) *AuthService {
	return &AuthService{users: users, tokens: tokens}
}

// Register creates a new account and returns the created user together
// with a signed token so the client can be authenticated immediately.
func (s *AuthService) Register(ctx context.Context, name, email, password string) (domain.User, string, error) {
	user := domain.User{Name: name, Email: email}
	user.Normalize()

	if err := validateName(user.Name); err != nil {
		return domain.User{}, "", err
	}
	if err := validateEmail(user.Email); err != nil {
		return domain.User{}, "", err
	}
	if err := validatePassword(password); err != nil {
		return domain.User{}, "", err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return domain.User{}, "", fmt.Errorf("hash password: %w", err)
	}

	created, err := s.users.Create(ctx, user.Name, user.Email, string(hash))
	if err != nil {
		return domain.User{}, "", err
	}

	signed, err := s.tokens.Issue(created.ID)
	if err != nil {
		return domain.User{}, "", fmt.Errorf("issue token: %w", err)
	}
	return created, signed, nil
}

// Login verifies the credentials and issues a token. The same error is
// returned whether the email or the password is wrong, to avoid leaking
// which accounts exist.
func (s *AuthService) Login(ctx context.Context, email, password string) (domain.User, string, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return domain.User{}, "", domain.ErrInvalidCredentials
		}
		return domain.User{}, "", err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return domain.User{}, "", domain.ErrInvalidCredentials
	}

	signed, err := s.tokens.Issue(user.ID)
	if err != nil {
		return domain.User{}, "", fmt.Errorf("issue token: %w", err)
	}
	return user, signed, nil
}

// Me returns the profile of the authenticated user.
func (s *AuthService) Me(ctx context.Context, userID int64) (domain.User, error) {
	return s.users.FindByID(ctx, userID)
}

func validateName(name string) error {
	if len(name) < minNameLength || len(name) > maxNameLength {
		return domain.ValidationError{
			Field:  "name",
			Reason: fmt.Sprintf("must be between %d and %d characters", minNameLength, maxNameLength),
		}
	}
	return nil
}

func validateEmail(email string) error {
	if !strings.Contains(email, "@") || len(email) > maxEmailLength {
		return domain.ValidationError{Field: "email", Reason: "must be a valid email address"}
	}
	return nil
}

func validatePassword(password string) error {
	if len(password) < minPasswordLength {
		return domain.ValidationError{
			Field:  "password",
			Reason: fmt.Sprintf("must be at least %d characters", minPasswordLength),
		}
	}
	return nil
}
