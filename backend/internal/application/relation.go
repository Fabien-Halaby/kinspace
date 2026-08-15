package application

import (
	"context"
	"strings"

	"github.com/Fabien-Halaby/kinspace/backend/internal/domain"
)

// RelationService implements the family-graph use cases.
type RelationService struct {
	repo domain.RelationRepository
}

func NewRelationService(repo domain.RelationRepository) *RelationService {
	return &RelationService{repo: repo}
}

// Create stores a directed relation edge. The relation is only allowed
// between members of the same family and never with oneself.
func (s *RelationService) Create(ctx context.Context, familyID, userID, relatedUserID int64, relationType string) (domain.Relation, error) {
	relationType = strings.ToLower(strings.TrimSpace(relationType))
	if !domain.IsValidRelationType(relationType) {
		return domain.Relation{}, domain.ErrInvalidRelationType
	}
	if userID == relatedUserID {
		return domain.Relation{}, domain.ErrSelfRelation
	}

	otherFamilyID, err := s.repo.UserFamilyID(ctx, relatedUserID)
	if err != nil {
		return domain.Relation{}, err
	}
	if otherFamilyID != familyID {
		return domain.Relation{}, domain.ErrNotInSameFamily
	}
	return s.repo.Create(ctx, familyID, userID, relatedUserID, relationType)
}

// List returns every relation edge stored for the family.
func (s *RelationService) List(ctx context.Context, familyID int64) ([]domain.Relation, error) {
	return s.repo.List(ctx, familyID)
}
