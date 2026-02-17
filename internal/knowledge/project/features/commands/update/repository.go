package update

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	projdomain "documind.jordi.org/internal/knowledge/domain"
	shared "documind.jordi.org/internal/shared/domain"
)

type Repository interface {
	GetByID(ctx context.Context, id shared.ID) (*projdomain.Project, error)
	Update(ctx context.Context, project *projdomain.Project) error
}

type postgresRepo struct{ db *sql.DB }

func newPostgresRepo(db *sql.DB) *postgresRepo { return &postgresRepo{db: db} }

func (r *postgresRepo) GetByID(ctx context.Context, id shared.ID) (*projdomain.Project, error) {
	query := `
		SELECT id, workspace_id, name, slug, description, created_at, updated_at, created_by, updated_by
		FROM projects WHERE id = $1
	`
	var project projdomain.Project
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&project.ID, &project.WorkspaceID, &project.Name, &project.Slug, &project.Description,
		&project.CreatedAt, &project.UpdatedAt, &project.CreatedBy, &project.UpdatedBy,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	return &project, nil
}

func (r *postgresRepo) Update(ctx context.Context, project *projdomain.Project) error {
	query := `
		UPDATE projects SET name = $2, description = $3, updated_at = $4, updated_by = $5 WHERE id = $1
	`
	result, err := r.db.ExecContext(ctx, query, project.ID, project.Name, project.Description, project.UpdatedAt, project.UpdatedBy)
	if err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return errors.New("project not found")
	}
	return nil
}
