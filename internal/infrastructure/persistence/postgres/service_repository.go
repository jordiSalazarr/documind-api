package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"documind.jordi.org/internal/domain/common"
	"documind.jordi.org/internal/domain/knowledge"
)

type ServiceRepository struct {
	db *sql.DB
}

func NewServiceRepository(db *sql.DB) *ServiceRepository {
	return &ServiceRepository{db: db}
}

func (r *ServiceRepository) Create(ctx context.Context, service *knowledge.Service) error {
	query := `
		INSERT INTO services (
			id, workspace_id, project_id, name, slug, description,
			created_at, updated_at, created_by, updated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.db.ExecContext(ctx, query,
		service.ID, service.WorkspaceID, service.ProjectID, service.Name, service.Slug, service.Description,
		service.CreatedAt, service.UpdatedAt, service.CreatedBy, service.UpdatedBy,
	)

	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}

	return nil
}

func (r *ServiceRepository) GetByID(ctx context.Context, id common.ID) (*knowledge.Service, error) {
	query := `
		SELECT id, workspace_id, project_id, name, slug, description,
			created_at, updated_at, created_by, updated_by
		FROM services
		WHERE id = $1
	`

	var service knowledge.Service
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&service.ID, &service.WorkspaceID, &service.ProjectID, &service.Name, &service.Slug, &service.Description,
		&service.CreatedAt, &service.UpdatedAt, &service.CreatedBy, &service.UpdatedBy,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("service not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get service: %w", err)
	}

	return &service, nil
}

func (r *ServiceRepository) GetBySlug(ctx context.Context, projectID common.ID, slug common.Slug) (*knowledge.Service, error) {
	query := `
		SELECT id, workspace_id, project_id, name, slug, description,
			created_at, updated_at, created_by, updated_by
		FROM services
		WHERE project_id = $1 AND slug = $2
	`

	var service knowledge.Service
	err := r.db.QueryRowContext(ctx, query, projectID, slug).Scan(
		&service.ID, &service.WorkspaceID, &service.ProjectID, &service.Name, &service.Slug, &service.Description,
		&service.CreatedAt, &service.UpdatedAt, &service.CreatedBy, &service.UpdatedBy,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("service not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get service by slug: %w", err)
	}

	return &service, nil
}

func (r *ServiceRepository) ListByProject(ctx context.Context, projectID common.ID) ([]*knowledge.Service, error) {
	query := `
		SELECT id, workspace_id, project_id, name, slug, description,
			created_at, updated_at, created_by, updated_by
		FROM services
		WHERE project_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}
	defer rows.Close()

	var services []*knowledge.Service
	for rows.Next() {
		var service knowledge.Service
		err := rows.Scan(
			&service.ID, &service.WorkspaceID, &service.ProjectID, &service.Name, &service.Slug, &service.Description,
			&service.CreatedAt, &service.UpdatedAt, &service.CreatedBy, &service.UpdatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan service: %w", err)
		}
		services = append(services, &service)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating services: %w", err)
	}

	return services, nil
}

func (r *ServiceRepository) Update(ctx context.Context, service *knowledge.Service) error {
	query := `
		UPDATE services
		SET name = $2, description = $3, updated_at = $4, updated_by = $5
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		service.ID, service.Name, service.Description, service.UpdatedAt, service.UpdatedBy,
	)

	if err != nil {
		return fmt.Errorf("failed to update service: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return errors.New("service not found")
	}

	return nil
}

func (r *ServiceRepository) Delete(ctx context.Context, id common.ID) error {
	query := `DELETE FROM services WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete service: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return errors.New("service not found")
	}

	return nil
}

func (r *ServiceRepository) Exists(ctx context.Context, id common.ID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM services WHERE id = $1)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check service existence: %w", err)
	}

	return exists, nil
}
