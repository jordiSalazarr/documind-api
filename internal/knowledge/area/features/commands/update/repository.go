package update

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	svcdomain "documind.jordi.org/internal/knowledge/domain"
	shared "documind.jordi.org/internal/shared/domain"
)

type Repository interface {
	GetByID(ctx context.Context, id shared.ID) (*svcdomain.Area, error)
	Update(ctx context.Context, area *svcdomain.Area) error
}

type postgresRepo struct{ db *sql.DB }

func newPostgresRepo(db *sql.DB) *postgresRepo { return &postgresRepo{db: db} }

func (r *postgresRepo) GetByID(ctx context.Context, id shared.ID) (*svcdomain.Area, error) {
	query := `
		SELECT id, project_id, name, slug, description, created_at, updated_at
		FROM services WHERE id = $1
	`
	var area svcdomain.Area
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&area.ID, &area.ProjectID, &area.Name, &area.Slug, &area.Description,
		&area.CreatedAt, &area.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get area: %w", err)
	}
	return &area, nil
}

func (r *postgresRepo) Update(ctx context.Context, area *svcdomain.Area) error {
	query := `UPDATE services SET name = $2, description = $3, updated_at = $4 WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, area.ID, area.Name, area.Description, area.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to update area: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return errors.New("area not found")
	}
	return nil
}
