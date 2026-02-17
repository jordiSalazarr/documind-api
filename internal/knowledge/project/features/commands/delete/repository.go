package delete

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
	Delete(ctx context.Context, id shared.ID) error
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

func (r *postgresRepo) Delete(ctx context.Context, id shared.ID) error {
	query := `DELETE FROM projects WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
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
