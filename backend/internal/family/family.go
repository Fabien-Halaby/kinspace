package family

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
)

var ErrAlreadyInFamily = errors.New("user already belongs to a family")
var ErrFamilyNotFound = errors.New("family not found")
var ErrInviteCodeInvalid = errors.New("invalid invite code")

type Family struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	InviteCode string `json:"invite_code"`
}

type Repository interface {
	Create(ctx context.Context, name, inviteCode string, ownerID int64) (Family, error)
	FindByUserID(ctx context.Context, userID int64) (Family, error)
	Join(ctx context.Context, userID int64, inviteCode string) (Family, error)
}

type Service struct{ repo Repository }
func NewService(repo Repository) *Service { return &Service{repo: repo} }

func (s *Service) Create(ctx context.Context, userID int64, name string) (Family, error) {
	name = strings.TrimSpace(name)
	if len(name) < 2 || len(name) > 120 { return Family{}, fmt.Errorf("family name must be between 2 and 120 characters") }
	if _, err := s.repo.FindByUserID(ctx, userID); err == nil { return Family{}, ErrAlreadyInFamily }
	code, err := generateInviteCode()
	if err != nil { return Family{}, fmt.Errorf("generate invite code: %w", err) }
	return s.repo.Create(ctx, name, code, userID)
}

func (s *Service) Join(ctx context.Context, userID int64, inviteCode string) (Family, error) {
	if _, err := s.repo.FindByUserID(ctx, userID); err == nil { return Family{}, ErrAlreadyInFamily }
	return s.repo.Join(ctx, userID, strings.TrimSpace(inviteCode))
}

func generateInviteCode() (string, error) {
	buf := make([]byte, 9)
	if _, err := rand.Read(buf); err != nil { return "", err }
	return strings.ToUpper(base64.RawURLEncoding.EncodeToString(buf)), nil
}
