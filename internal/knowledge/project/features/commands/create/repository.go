package create

import (
	"context"
	"database/sql"
	"fmt"

	projdomain "documind.jordi.org/internal/knowledge/domain"
	shared "documind.jordi.org/internal/shared/domain"
)

type Repository interface {
	Insert(ctx context.Context, project *projdomain.Project) error
	GetBySlug(ctx context.Context, workspaceID shared.ID, slug shared.Slug) (*projdomain.Project, error)
}

type postgresRepo struct{ db *sql.DB }

func newPostgresRepo(db *sql.DB) *postgresRepo { return &postgresRepo{db: db} }

func (r *postgresRepo) Insert(ctx context.Context, project *projdomain.Project) error {
	query := `
		INSERT INTO projects (id, workspace_id, name, slug, description, created_at, updated_at, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.ExecContext(ctx, query,
		project.ID, project.WorkspaceID, project.Name, project.Slug, project.Description,
		project.CreatedAt, project.UpdatedAt, project.CreatedBy, project.UpdatedBy,
	)
	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}
	return nil
}

func (r *postgresRepo) GetBySlug(ctx context.Context, workspaceID shared.ID, slug shared.Slug) (*projdomain.Project, error) {
	query := `
		SELECT id, workspace_id, name, slug, description, created_at, updated_at, created_by, updated_by
		FROM projects WHERE workspace_id = $1 AND slug = $2
	`
	var project projdomain.Project
	err := r.db.QueryRowContext(ctx, query, workspaceID, slug).Scan(
		&project.ID, &project.WorkspaceID, &project.Name, &project.Slug, &project.Description,
		&project.CreatedAt, &project.UpdatedAt, &project.CreatedBy, &project.UpdatedBy,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get project by slug: %w", err)
	}
	return &project, nil
}
