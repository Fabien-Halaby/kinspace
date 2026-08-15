package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Fabien-Halaby/kinspace/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// FamilyRepository implements domain.FamilyRepository on PostgreSQL.
type FamilyRepository struct {
	db *pgxpool.Pool
}

func NewFamilyRepository(db *pgxpool.Pool) *FamilyRepository {
	return &FamilyRepository{db: db}
}

// Create inserts the family and assigns its owner in a single
// transaction so a partial state (family without owner) can never be
// committed.
func (r *FamilyRepository) Create(ctx context.Context, name, inviteCode string, ownerID int64) (domain.Family, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return domain.Family{}, fmt.Errorf("begin family transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var family domain.Family
	err = tx.QueryRow(ctx, `
		INSERT INTO families (name, invite_code)
		VALUES ($1, $2)
		RETURNING id, name, invite_code, created_at
	`, name, inviteCode).Scan(&family.ID, &family.Name, &family.InviteCode, &family.CreatedAt)
	if err != nil {
		return domain.Family{}, fmt.Errorf("insert family: %w", err)
	}

	result, err := tx.Exec(ctx, `
		UPDATE users
		SET family_id = $1
		WHERE id = $2 AND family_id IS NULL
	`, family.ID, ownerID)
	if err != nil {
		return domain.Family{}, fmt.Errorf("assign family owner: %w", err)
	}
	if result.RowsAffected() != 1 {
		return domain.Family{}, domain.ErrAlreadyInFamily
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Family{}, fmt.Errorf("commit family transaction: %w", err)
	}
	return family, nil
}

func (r *FamilyRepository) FindByUserID(ctx context.Context, userID int64) (domain.Family, error) {
	var family domain.Family
	err := r.db.QueryRow(ctx, `
		SELECT f.id, f.name, f.invite_code, f.created_at
		FROM families f
		JOIN users u ON u.family_id = f.id
		WHERE u.id = $1
	`, userID).Scan(&family.ID, &family.Name, &family.InviteCode, &family.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Family{}, domain.ErrFamilyNotFound
	}
	if err != nil {
		return domain.Family{}, fmt.Errorf("find family by user: %w", err)
	}
	return family, nil
}

func (r *FamilyRepository) Join(ctx context.Context, userID int64, inviteCode string) (domain.Family, error) {
	var family domain.Family
	err := r.db.QueryRow(ctx, `
		SELECT id, name, invite_code, created_at
		FROM families
		WHERE invite_code = $1
	`, inviteCode).Scan(&family.ID, &family.Name, &family.InviteCode, &family.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Family{}, domain.ErrInviteCodeInvalid
	}
	if err != nil {
		return domain.Family{}, fmt.Errorf("find family by invite code: %w", err)
	}

	result, err := r.db.Exec(ctx, `
		UPDATE users
		SET family_id = $1
		WHERE id = $2 AND family_id IS NULL
	`, family.ID, userID)
	if err != nil {
		return domain.Family{}, fmt.Errorf("join family: %w", err)
	}
	if result.RowsAffected() != 1 {
		return domain.Family{}, domain.ErrAlreadyInFamily
	}
	return family, nil
}
