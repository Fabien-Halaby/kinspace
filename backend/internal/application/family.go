package application

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/Fabien-Halaby/kinspace/backend/internal/domain"
)

const (
	minFamilyNameLength = 2
	maxFamilyNameLength = 120
)

// FamilyService implements the family use cases: lookup, creation and
// joining through an invite code.
type FamilyService struct {
	repo domain.FamilyRepository
}

func NewFamilyService(repo domain.FamilyRepository) *FamilyService {
	return &FamilyService{repo: repo}
}

// Me returns the family the user currently belongs to.
func (s *FamilyService) Me(ctx context.Context, userID int64) (domain.Family, error) {
	return s.repo.FindByUserID(ctx, userID)
}

// Create makes the user the owner of a brand-new family.
func (s *FamilyService) Create(ctx context.Context, userID int64, name string) (domain.Family, error) {
	name = strings.TrimSpace(name)
	if len(name) < minFamilyNameLength || len(name) > maxFamilyNameLength {
		return domain.Family{}, domain.ValidationError{
			Field:  "name",
			Reason: fmt.Sprintf("must be between %d and %d characters", minFamilyNameLength, maxFamilyNameLength),
		}
	}
	if err := s.ensureNoFamily(ctx, userID); err != nil {
		return domain.Family{}, err
	}

	code, err := generateInviteCode()
	if err != nil {
		return domain.Family{}, fmt.Errorf("generate invite code: %w", err)
	}
	return s.repo.Create(ctx, name, code, userID)
}

// Join attaches the user to the family matching the invite code.
func (s *FamilyService) Join(ctx context.Context, userID int64, inviteCode string) (domain.Family, error) {
	inviteCode = strings.TrimSpace(inviteCode)
	if inviteCode == "" {
		return domain.Family{}, domain.ValidationError{Field: "invite_code", Reason: "is required"}
	}
	if err := s.ensureNoFamily(ctx, userID); err != nil {
		return domain.Family{}, err
	}
	return s.repo.Join(ctx, userID, inviteCode)
}

// ensureNoFamily short-circuits users that already belong to a family.
func (s *FamilyService) ensureNoFamily(ctx context.Context, userID int64) error {
	if _, err := s.repo.FindByUserID(ctx, userID); err == nil {
		return domain.ErrAlreadyInFamily
	} else if !errors.Is(err, domain.ErrFamilyNotFound) {
		return err
	}
	return nil
}

func generateInviteCode() (string, error) {
	buf := make([]byte, 9)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return strings.ToUpper(base64.RawURLEncoding.EncodeToString(buf)), nil
}
