package token

import (
	"encoding/base64"
	"errors"
	"strings"
	"testing"
	"time"
)

func newManager() *HS256Manager {
	return NewHS256Manager("this-is-a-secret-with-at-least-32-chars", time.Hour)
}

func TestIssueParseRoundTrip(t *testing.T) {
	m := newManager()
	raw, err := m.Issue(42)
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	userID, err := m.Parse(raw)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if userID != 42 {
		t.Fatalf("userID = %d, want 42", userID)
	}
}

func TestParseRejectsGarbage(t *testing.T) {
	m := newManager()
	if _, err := m.Parse("not.a.jwt"); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("error = %v, want ErrInvalidToken", err)
	}
}

func TestParseRejectsWrongSignature(t *testing.T) {
	m := newManager()
	other := NewHS256Manager("a-different-secret-that-is-also-32-chars", time.Hour)
	raw, err := other.Issue(1)
	if err != nil {
		t.Fatalf("seed issue: %v", err)
	}

	if _, err := m.Parse(raw); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("error = %v, want ErrInvalidToken", err)
	}
}

func TestParseRejectsExpiredToken(t *testing.T) {
	m := NewHS256Manager("this-is-a-secret-with-at-least-32-chars", -time.Hour)
	raw, err := m.Issue(1)
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	if _, err := m.Parse(raw); !errors.Is(err, ErrExpiredToken) {
		t.Fatalf("error = %v, want ErrExpiredToken", err)
	}
}

func TestParseRejectsAlgNone(t *testing.T) {
	m := newManager()
	raw, err := m.Issue(1)
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	// Rebuild the token with an unsigned "none" header to prove that
	// algorithm confusion is rejected.
	parts := strings.Split(raw, ".")
	if len(parts) != 3 {
		t.Fatalf("token has %d parts, want 3", len(parts))
	}
	tampered := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none","typ":"JWT"}`)) +
		"." + parts[1] + "." + parts[2]

	if _, err := m.Parse(tampered); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("error = %v, want ErrInvalidToken", err)
	}
}
