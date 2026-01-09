package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"documind.jordi.org/internal/domain/common"
	"documind.jordi.org/internal/domain/knowledge"
)

type ProjectRepository struct {
	db *sql.DB
}

func NewProjectRepository(db *sql.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) Create(ctx context.Context, project *knowledge.Project) error {
	query := `
		INSERT INTO projects (
			id, workspace_id, name, slug, description,
			created_at, updated_at, created_by, updated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
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

func (r *ProjectRepository) GetByID(ctx context.Context, id common.ID) (*knowledge.Project, error) {
	query := `
		SELECT id, workspace_id, name, slug, description,
			created_at, updated_at, created_by, updated_by
		FROM projects
		WHERE id = $1
	`

	var project knowledge.Project
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&project.ID, &project.WorkspaceID, &project.Name, &project.Slug, &project.Description,
		&project.CreatedAt, &project.UpdatedAt, &project.CreatedBy, &project.UpdatedBy,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("project not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	return &project, nil
}

func (r *ProjectRepository) GetBySlug(ctx context.Context, workspaceID common.ID, slug common.Slug) (*knowledge.Project, error) {
	query := `
		SELECT id, workspace_id, name, slug, description,
			created_at, updated_at, created_by, updated_by
		FROM projects
		WHERE workspace_id = $1 AND slug = $2
	`

	var project knowledge.Project
	err := r.db.QueryRowContext(ctx, query, workspaceID, slug).Scan(
		&project.ID, &project.WorkspaceID, &project.Name, &project.Slug, &project.Description,
		&project.CreatedAt, &project.UpdatedAt, &project.CreatedBy, &project.UpdatedBy,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("project not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get project by slug: %w", err)
	}

	return &project, nil
}

func (r *ProjectRepository) ListByWorkspace(ctx context.Context, workspaceID common.ID) ([]*knowledge.Project, error) {
	query := `
		SELECT id, workspace_id, name, slug, description,
			created_at, updated_at, created_by, updated_by
		FROM projects
		WHERE workspace_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}
	defer rows.Close()

	var projects []*knowledge.Project
	for rows.Next() {
		var project knowledge.Project
		err := rows.Scan(
			&project.ID, &project.WorkspaceID, &project.Name, &project.Slug, &project.Description,
			&project.CreatedAt, &project.UpdatedAt, &project.CreatedBy, &project.UpdatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}
		projects = append(projects, &project)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating projects: %w", err)
	}

	return projects, nil
}

func (r *ProjectRepository) Update(ctx context.Context, project *knowledge.Project) error {
	query := `
		UPDATE projects
		SET name = $2, description = $3, updated_at = $4, updated_by = $5
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		project.ID, project.Name, project.Description, project.UpdatedAt, project.UpdatedBy,
	)

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

func (r *ProjectRepository) Delete(ctx context.Context, id common.ID) error {
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

func (r *ProjectRepository) Exists(ctx context.Context, id common.ID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM projects WHERE id = $1)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check project existence: %w", err)
	}

	return exists, nil
}
