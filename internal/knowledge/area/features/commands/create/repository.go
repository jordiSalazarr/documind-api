package create

import (
	"context"
	"database/sql"
	"fmt"

	svcdomain "documind.jordi.org/internal/knowledge/domain"
	shared "documind.jordi.org/internal/shared/domain"
)

type Repository interface {
	Insert(ctx context.Context, area *svcdomain.Area) error
	GetBySlug(ctx context.Context, projectID shared.ID, slug shared.Slug) (*svcdomain.Area, error)
	ProjectExists(ctx context.Context, id shared.ID) (bool, error)
}

type postgresRepo struct{ db *sql.DB }

func newPostgresRepo(db *sql.DB) *postgresRepo { return &postgresRepo{db: db} }

func (r *postgresRepo) Insert(ctx context.Context, area *svcdomain.Area) error {
	query := `
		INSERT INTO services (id, project_id, name, slug, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.ExecContext(ctx, query,
		area.ID, area.ProjectID, area.Name, area.Slug, area.Description,
		area.CreatedAt, area.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create area: %w", err)
	}
	return nil
}

func (r *postgresRepo) GetBySlug(ctx context.Context, projectID shared.ID, slug shared.Slug) (*svcdomain.Area, error) {
	query := `
		SELECT id, project_id, name, slug, description, created_at, updated_at
		FROM services WHERE project_id = $1 AND slug = $2
	`
	var area svcdomain.Area
	err := r.db.QueryRowContext(ctx, query, projectID, slug).Scan(
		&area.ID, &area.ProjectID, &area.Name, &area.Slug, &area.Description,
		&area.CreatedAt, &area.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get area by slug: %w", err)
	}
	return &area, nil
}

func (r *postgresRepo) ProjectExists(ctx context.Context, id shared.ID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM projects WHERE id = $1)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check project existence: %w", err)
	}
	return exists, nil
}
