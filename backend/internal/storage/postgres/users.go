package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Fabien-Halaby/kinspace/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UserRepository implements domain.UserRepository on PostgreSQL.
type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, name, email, passwordHash string) (domain.User, error) {
	var user domain.User
	err := r.db.QueryRow(ctx, `
		INSERT INTO users (name, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, name, email, family_id, created_at
	`, name, email, passwordHash).Scan(
		&user.ID, &user.Name, &user.Email, &user.FamilyID, &user.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.User{}, domain.ErrEmailExists
		}
		return domain.User{}, fmt.Errorf("insert user: %w", err)
	}
	return user, nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	var user domain.User
	err := r.db.QueryRow(ctx, `
		SELECT id, name, email, password_hash, family_id, created_at
		FROM users
		WHERE email = $1
	`, email).Scan(
		&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.FamilyID, &user.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, domain.ErrUserNotFound
	}
	if err != nil {
		return domain.User{}, fmt.Errorf("find user by email: %w", err)
	}
	return user, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id int64) (domain.User, error) {
	var user domain.User
	err := r.db.QueryRow(ctx, `
		SELECT id, name, email, password_hash, family_id, created_at
		FROM users
		WHERE id = $1
	`, id).Scan(
		&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.FamilyID, &user.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, domain.ErrUserNotFound
	}
	if err != nil {
		return domain.User{}, fmt.Errorf("find user by id: %w", err)
	}
	return user, nil
}
