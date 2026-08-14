package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrInvalidToken = errors.New("invalid token")

type TokenManager struct {
	secret []byte
	ttl    time.Duration
}

type claims struct {
	Sub int64 `json:"sub"`
	Exp int64 `json:"exp"`
	Iat int64 `json:"iat"`
}

func NewTokenManager(secret string, ttl time.Duration) (*TokenManager, error) {
	if len(secret) < 32 { return nil, errors.New("JWT secret must be at least 32 characters") }
	if ttl <= 0 { return nil, errors.New("JWT TTL must be positive") }
	return &TokenManager{secret: []byte(secret), ttl: ttl}, nil
}

func (m *TokenManager) Issue(user User) (string, error) {
	header, err := json.Marshal(map[string]string{"alg":"HS256", "typ":"JWT"})
	if err != nil { return "", fmt.Errorf("encode header: %w", err) }
	now := time.Now().Unix()
	body, err := json.Marshal(claims{Sub:user.ID, Iat:now, Exp:time.Now().Add(m.ttl).Unix()})
	if err != nil { return "", fmt.Errorf("encode claims: %w", err) }
	encoded := base64.RawURLEncoding.EncodeToString(header)+"."+base64.RawURLEncoding.EncodeToString(body)
	return encoded+"."+m.sign(encoded), nil
}

func (m *TokenManager) Parse(token string) (int64, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 || !hmac.Equal([]byte(parts[2]), []byte(m.sign(parts[0]+"."+parts[1]))) { return 0, ErrInvalidToken }
	payload, err := base64.RawURLEncoding.DecodeString(parts[1]); if err != nil { return 0, ErrInvalidToken }
	var c claims
	if err := json.Unmarshal(payload, &c); err != nil || c.Sub <= 0 || c.Exp <= time.Now().Unix() { return 0, ErrInvalidToken }
	return c.Sub, nil
}

func (m *TokenManager) sign(input string) string {
	h := hmac.New(sha256.New, m.secret); _, _ = h.Write([]byte(input)); return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}
