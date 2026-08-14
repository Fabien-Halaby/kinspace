package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresRepository struct{ db *pgxpool.Pool }

func NewPostgresRepository(db *pgxpool.Pool) Repository { return &postgresRepository{db: db} }

func (r *postgresRepository) Create(ctx context.Context, name, email, passwordHash string) (User, error) {
	var user User
	err := r.db.QueryRow(ctx, `
		INSERT INTO users (name, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, name, email, family_id
	`, name, email, passwordHash).Scan(&user.ID, &user.Name, &user.Email, &user.FamilyID)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return User{}, ErrEmailExists
		}
		return User{}, fmt.Errorf("create user: %w", err)
	}
	return user, nil
}

func (r *postgresRepository) FindByEmail(ctx context.Context, email string) (User, error) {
	var user User
	err := r.db.QueryRow(ctx, `
		SELECT id, name, email, password_hash, family_id
		FROM users WHERE email = $1
	`, email).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.FamilyID)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrInvalidCredentials
	}
	if err != nil {
		return User{}, fmt.Errorf("find user: %w", err)
	}
	return user, nil
}

func RegisterHandler(service *Service) gin.HandlerFunc {
	type request struct {
		Name string `json:"name" binding:"required"`
		Email string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=8"`
	}
	return func(c *gin.Context) {
		var req request
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
			return
		}
		user, err := service.Register(c.Request.Context(), req.Name, req.Email, req.Password)
		if errors.Is(err, ErrEmailExists) {
			c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
			return
		}
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"user": user})
	}
}

func LoginHandler(service *Service, issueToken func(User) (string, error)) gin.HandlerFunc {
	type request struct {
		Email string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}
	return func(c *gin.Context) {
		var req request
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
			return
		}
		user, err := service.Login(c.Request.Context(), req.Email, req.Password)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}
		token, err := issueToken(user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not issue token"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"token": token, "user": user})
	}
}
