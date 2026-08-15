package token

import (
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

// Manager is the port used by the application layer to issue and verify
// signed tokens. Keeping it as an interface keeps the JWT implementation
// swappable and the use cases testable.
type Manager interface {
	Issue(userID int64) (string, error)
	Parse(token string) (int64, error)
}

const issuer = "kinspace-api"

// HS256Manager signs and verifies HS256 JWTs.
type HS256Manager struct {
	secret []byte
	ttl    time.Duration
}

func NewHS256Manager(secret string, ttl time.Duration) *HS256Manager {
	return &HS256Manager{secret: []byte(secret), ttl: ttl}
}

// Issue signs a token carrying the user identifier as the subject.
func (m *HS256Manager) Issue(userID int64) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Issuer:    issuer,
		Subject:   strconv.FormatInt(userID, 10),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
}

// Parse verifies the signature, validity window and signing algorithm of
// a token and returns the subject as the user identifier.
func (m *HS256Manager) Parse(raw string) (int64, error) {
	claims := &jwt.RegisteredClaims{}
	_, err := jwt.ParseWithClaims(
		raw,
		claims,
		func(*jwt.Token) (any, error) { return m.secret, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(issuer),
	)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return 0, ErrExpiredToken
		}
		return 0, ErrInvalidToken
	}

	subject, err := claims.GetSubject()
	if err != nil {
		return 0, ErrInvalidToken
	}
	userID, err := strconv.ParseInt(subject, 10, 64)
	if err != nil || userID <= 0 {
		return 0, ErrInvalidToken
	}
	return userID, nil
}
