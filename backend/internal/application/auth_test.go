package application

import (
	"context"
	"errors"
	"testing"

	"github.com/Fabien-Halaby/kinspace/backend/internal/domain"
)

type fakeUserRepo struct {
	users     map[string]domain.User
	created   domain.User
	nextID    int64
	createErr error
	findErr   error
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{users: make(map[string]domain.User), nextID: 1}
}

func (f *fakeUserRepo) Create(_ context.Context, name, email, passwordHash string) (domain.User, error) {
	if f.createErr != nil {
		return domain.User{}, f.createErr
	}
	f.created = domain.User{ID: f.nextID, Name: name, Email: email, PasswordHash: passwordHash}
	f.nextID++
	f.users[email] = f.created
	return f.created, nil
}

func (f *fakeUserRepo) FindByEmail(_ context.Context, email string) (domain.User, error) {
	if f.findErr != nil {
		return domain.User{}, f.findErr
	}
	if user, ok := f.users[email]; ok {
		return user, nil
	}
	return domain.User{}, domain.ErrUserNotFound
}

func (f *fakeUserRepo) FindByID(_ context.Context, id int64) (domain.User, error) {
	for _, user := range f.users {
		if user.ID == id {
			return user, nil
		}
	}
	return domain.User{}, domain.ErrUserNotFound
}

type fakeTokenManager struct {
	subject int64
}

func (f *fakeTokenManager) Issue(userID int64) (string, error) {
	f.subject = userID
	return "signed-token", nil
}

func (f *fakeTokenManager) Parse(string) (int64, error) {
	return f.subject, nil
}

func TestRegisterNormalizesEmailAndHashesPassword(t *testing.T) {
	repo := newFakeUserRepo()
	service := NewAuthService(repo, &fakeTokenManager{})

	user, signed, err := service.Register(context.Background(), " Fabien ", "FABIEN@Example.COM", "correct-password")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if user.Email != "fabien@example.com" {
		t.Fatalf("email = %q, want normalized email", user.Email)
	}
	if user.Name != "Fabien" {
		t.Fatalf("name = %q, want trimmed name", user.Name)
	}
	if repo.created.PasswordHash == "" || repo.created.PasswordHash == "correct-password" {
		t.Fatal("password was not hashed")
	}
	if signed == "" {
		t.Fatal("expected a token to be issued")
	}
}

func TestRegisterRejectsShortPassword(t *testing.T) {
	service := NewAuthService(newFakeUserRepo(), &fakeTokenManager{})
	_, _, err := service.Register(context.Background(), "Fabien", "fabien@example.com", "short")
	if err == nil {
		t.Fatal("expected short password to be rejected")
	}
}

func TestRegisterRejectsInvalidEmail(t *testing.T) {
	service := NewAuthService(newFakeUserRepo(), &fakeTokenManager{})
	_, _, err := service.Register(context.Background(), "Fabien", "not-an-email", "correct-password")
	if err == nil {
		t.Fatal("expected invalid email to be rejected")
	}
}

func TestRegisterPropagatesDuplicateEmail(t *testing.T) {
	repo := newFakeUserRepo()
	repo.createErr = domain.ErrEmailExists
	service := NewAuthService(repo, &fakeTokenManager{})

	_, _, err := service.Register(context.Background(), "Fabien", "fabien@example.com", "correct-password")
	if !errors.Is(err, domain.ErrEmailExists) {
		t.Fatalf("error = %v, want ErrEmailExists", err)
	}
}

func TestLoginRejectsUnknownEmail(t *testing.T) {
	service := NewAuthService(newFakeUserRepo(), &fakeTokenManager{})
	_, _, err := service.Login(context.Background(), "nobody@example.com", "whatever-password")
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("error = %v, want ErrInvalidCredentials", err)
	}
}

func TestLoginRejectsWrongPassword(t *testing.T) {
	repo := newFakeUserRepo()
	service := NewAuthService(repo, &fakeTokenManager{})
	if _, _, err := service.Register(context.Background(), "Fabien", "fabien@example.com", "correct-password"); err != nil {
		t.Fatalf("seed register: %v", err)
	}

	_, _, err := service.Login(context.Background(), "fabien@example.com", "wrong-password")
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("error = %v, want ErrInvalidCredentials", err)
	}
}

func TestLoginSucceedsWithCorrectCredentials(t *testing.T) {
	repo := newFakeUserRepo()
	service := NewAuthService(repo, &fakeTokenManager{})
	created, _, err := service.Register(context.Background(), "Fabien", "fabien@example.com", "correct-password")
	if err != nil {
		t.Fatalf("seed register: %v", err)
	}

	user, signed, err := service.Login(context.Background(), "fabien@example.com", "correct-password")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if user.ID != created.ID {
		t.Fatalf("user.ID = %d, want %d", user.ID, created.ID)
	}
	if signed == "" {
		t.Fatal("expected a token to be issued")
	}
}
