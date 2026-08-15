package application

import (
	"context"
	"errors"
	"testing"

	"github.com/Fabien-Halaby/kinspace/backend/internal/domain"
)

type fakeFamilyRepo struct {
	familyByUser map[int64]domain.Family
	inviteCode   string
	nextID       int64
}

func newFakeFamilyRepo() *fakeFamilyRepo {
	return &fakeFamilyRepo{familyByUser: make(map[int64]domain.Family), nextID: 1}
}

func (f *fakeFamilyRepo) Create(_ context.Context, name, inviteCode string, ownerID int64) (domain.Family, error) {
	if f.inviteCode != "" {
		return domain.Family{}, domain.ErrAlreadyInFamily
	}
	family := domain.Family{ID: f.nextID, Name: name, InviteCode: inviteCode}
	f.nextID++
	f.familyByUser[ownerID] = family
	return family, nil
}

func (f *fakeFamilyRepo) FindByUserID(_ context.Context, userID int64) (domain.Family, error) {
	family, ok := f.familyByUser[userID]
	if !ok {
		return domain.Family{}, domain.ErrFamilyNotFound
	}
	return family, nil
}

func (f *fakeFamilyRepo) Join(_ context.Context, userID int64, inviteCode string) (domain.Family, error) {
	if inviteCode == "UNKNOWN" {
		return domain.Family{}, domain.ErrInviteCodeInvalid
	}
	if _, ok := f.familyByUser[userID]; ok {
		return domain.Family{}, domain.ErrAlreadyInFamily
	}
	family := domain.Family{ID: 99, Name: "Joined Family", InviteCode: inviteCode}
	f.familyByUser[userID] = family
	return family, nil
}

func TestFamilyCreateStoresOwner(t *testing.T) {
	repo := newFakeFamilyRepo()
	service := NewFamilyService(repo)

	family, err := service.Create(context.Background(), 7, "The Smiths")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if family.Name != "The Smiths" {
		t.Fatalf("name = %q, want The Smiths", family.Name)
	}
	if family.InviteCode == "" {
		t.Fatal("expected an invite code to be generated")
	}
	if _, ok := repo.familyByUser[7]; !ok {
		t.Fatal("owner was not assigned to the family")
	}
}

func TestFamilyCreateRejectsShortName(t *testing.T) {
	service := NewFamilyService(newFakeFamilyRepo())
	_, err := service.Create(context.Background(), 7, "A")
	if err == nil {
		t.Fatal("expected short name to be rejected")
	}
}

func TestFamilyCreateRejectsExistingMembership(t *testing.T) {
	repo := newFakeFamilyRepo()
	service := NewFamilyService(repo)
	if _, err := service.Create(context.Background(), 7, "First"); err != nil {
		t.Fatalf("seed create: %v", err)
	}

	_, err := service.Create(context.Background(), 7, "Second")
	if !errors.Is(err, domain.ErrAlreadyInFamily) {
		t.Fatalf("error = %v, want ErrAlreadyInFamily", err)
	}
}

func TestFamilyJoinRejectsInvalidInviteCode(t *testing.T) {
	service := NewFamilyService(newFakeFamilyRepo())
	_, err := service.Join(context.Background(), 7, "UNKNOWN")
	if !errors.Is(err, domain.ErrInviteCodeInvalid) {
		t.Fatalf("error = %v, want ErrInviteCodeInvalid", err)
	}
}

func TestFamilyJoinRejectsEmptyInviteCode(t *testing.T) {
	service := NewFamilyService(newFakeFamilyRepo())
	_, err := service.Join(context.Background(), 7, "   ")
	if err == nil {
		t.Fatal("expected empty invite code to be rejected")
	}
}

func TestFamilyMe(t *testing.T) {
	repo := newFakeFamilyRepo()
	service := NewFamilyService(repo)
	if _, err := service.Create(context.Background(), 7, "The Smiths"); err != nil {
		t.Fatalf("seed create: %v", err)
	}

	family, err := service.Me(context.Background(), 7)
	if err != nil {
		t.Fatalf("Me() error = %v", err)
	}
	if family.Name != "The Smiths" {
		t.Fatalf("name = %q, want The Smiths", family.Name)
	}
}
