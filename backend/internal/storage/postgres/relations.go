package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Fabien-Halaby/kinspace/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RelationRepository implements domain.RelationRepository on PostgreSQL.
type RelationRepository struct {
	db *pgxpool.Pool
}

func NewRelationRepository(db *pgxpool.Pool) *RelationRepository {
	return &RelationRepository{db: db}
}

func (r *RelationRepository) Create(ctx context.Context, familyID, userID, relatedUserID int64, relationType string) (domain.Relation, error) {
	var relation domain.Relation
	err := r.db.QueryRow(ctx, `
		INSERT INTO relations (family_id, user_id, related_user_id, type)
		VALUES ($1, $2, $3, $4)
		RETURNING id, family_id, user_id, related_user_id, type, created_at
	`, familyID, userID, relatedUserID, relationType).Scan(
		&relation.ID, &relation.FamilyID, &relation.UserID,
		&relation.RelatedUserID, &relation.Type, &relation.CreatedAt,
	)
	if err != nil {
		return domain.Relation{}, fmt.Errorf("insert relation: %w", err)
	}
	return relation, nil
}

func (r *RelationRepository) List(ctx context.Context, familyID int64) ([]domain.Relation, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, family_id, user_id, related_user_id, type, created_at
		FROM relations
		WHERE family_id = $1
		ORDER BY id
	`, familyID)
	if err != nil {
		return nil, fmt.Errorf("list relations: %w", err)
	}
	defer rows.Close()

	relations := make([]domain.Relation, 0)
	for rows.Next() {
		var relation domain.Relation
		if err := rows.Scan(
			&relation.ID, &relation.FamilyID, &relation.UserID,
			&relation.RelatedUserID, &relation.Type, &relation.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan relation: %w", err)
		}
		relations = append(relations, relation)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate relations: %w", err)
	}
	return relations, nil
}

func (r *RelationRepository) UserFamilyID(ctx context.Context, userID int64) (int64, error) {
	var familyID int64
	err := r.db.QueryRow(ctx, `SELECT family_id FROM users WHERE id = $1`, userID).Scan(&familyID)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, domain.ErrUserNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("find user family: %w", err)
	}
	return familyID, nil
}
