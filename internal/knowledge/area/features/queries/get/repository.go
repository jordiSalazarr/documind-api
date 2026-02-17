package get

import (
	"context"
	"database/sql"
	"fmt"

	svcdomain "documind.jordi.org/internal/knowledge/domain"
	shared "documind.jordi.org/internal/shared/domain"
)

type Repository interface {
	GetByID(ctx context.Context, id shared.ID) (*svcdomain.Area, error)
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
