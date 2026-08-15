package application

import (
	"context"
	"errors"
	"testing"

	"github.com/Fabien-Halaby/kinspace/backend/internal/domain"
)

type fakeRelationRepo struct {
	userFamily map[int64]int64
	created    domain.Relation
	createErr  error
	listed     []domain.Relation
	listErr    error
}

func newFakeRelationRepo() *fakeRelationRepo {
	return &fakeRelationRepo{userFamily: make(map[int64]int64)}
}

func (f *fakeRelationRepo) Create(_ context.Context, familyID, userID, relatedUserID int64, relationType string) (domain.Relation, error) {
	if f.createErr != nil {
		return domain.Relation{}, f.createErr
	}
	f.created = domain.Relation{ID: 1, FamilyID: familyID, UserID: userID, RelatedUserID: relatedUserID, Type: relationType}
	return f.created, nil
}

func (f *fakeRelationRepo) List(context.Context, int64) ([]domain.Relation, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.listed, nil
}

func (f *fakeRelationRepo) UserFamilyID(_ context.Context, userID int64) (int64, error) {
	if familyID, ok := f.userFamily[userID]; ok {
		return familyID, nil
	}
	return 0, domain.ErrUserNotFound
}

func TestRelationCreateStoresEdge(t *testing.T) {
	repo := newFakeRelationRepo()
	repo.userFamily[1] = 10
	repo.userFamily[2] = 10
	service := NewRelationService(repo)

	relation, err := service.Create(context.Background(), 10, 1, 2, " Parent ")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if relation.Type != "parent" {
		t.Fatalf("type = %q, want normalized parent", relation.Type)
	}
}

func TestRelationCreateRejectsInvalidType(t *testing.T) {
	service := NewRelationService(newFakeRelationRepo())
	_, err := service.Create(context.Background(), 10, 1, 2, "enemy")
	if !errors.Is(err, domain.ErrInvalidRelationType) {
		t.Fatalf("error = %v, want ErrInvalidRelationType", err)
	}
}

func TestRelationCreateRejectsSelfRelation(t *testing.T) {
	service := NewRelationService(newFakeRelationRepo())
	_, err := service.Create(context.Background(), 10, 1, 1, "parent")
	if !errors.Is(err, domain.ErrSelfRelation) {
		t.Fatalf("error = %v, want ErrSelfRelation", err)
	}
}

func TestRelationCreateRejectsCrossFamily(t *testing.T) {
	repo := newFakeRelationRepo()
	repo.userFamily[1] = 10
	repo.userFamily[2] = 20
	service := NewRelationService(repo)

	_, err := service.Create(context.Background(), 10, 1, 2, "parent")
	if !errors.Is(err, domain.ErrNotInSameFamily) {
		t.Fatalf("error = %v, want ErrNotInSameFamily", err)
	}
}

func TestRelationCreatePropagatesUnknownUser(t *testing.T) {
	repo := newFakeRelationRepo()
	repo.userFamily[1] = 10
	service := NewRelationService(repo)

	_, err := service.Create(context.Background(), 10, 1, 2, "parent")
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("error = %v, want ErrUserNotFound", err)
	}
}
