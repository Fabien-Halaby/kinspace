package auth

import (
	"context"
	"testing"
)

type fakeRepo struct{ created User }

func (f *fakeRepo) Create(_ context.Context, name, email, passwordHash string) (User, error) {
	f.created = User{ID: 1, Name: name, Email: email, PasswordHash: passwordHash}
	return f.created, nil
}

func (f *fakeRepo) FindByEmail(context.Context, string) (User, error) { return User{}, ErrInvalidCredentials }

func TestRegisterHashesPassword(t *testing.T) {
	repo := &fakeRepo{}
	service := NewService(repo)

	user, err := service.Register(context.Background(), "Fabien", "FABIEN@Example.COM", "correct-password")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if user.Email != "fabien@example.com" {
		t.Fatalf("email = %q, want normalized email", user.Email)
	}
	if repo.created.PasswordHash == "correct-password" || repo.created.PasswordHash == "" {
		t.Fatal("password was not hashed")
	}
}

func TestRegisterRejectsShortPassword(t *testing.T) {
	service := NewService(&fakeRepo{})
	_, err := service.Register(context.Background(), "Fabien", "fabien@example.com", "short")
	if err == nil {
		t.Fatal("expected short password to be rejected")
	}
}
