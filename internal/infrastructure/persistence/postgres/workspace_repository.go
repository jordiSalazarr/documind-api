package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"documind.jordi.org/internal/domain/common"
	"documind.jordi.org/internal/domain/workspace"
)

type WorkspaceRepository struct {
	db *sql.DB
}

func NewWorkspaceRepository(db *sql.DB) *WorkspaceRepository {
	return &WorkspaceRepository{db: db}
}

func (r *WorkspaceRepository) Create(ctx context.Context, ws *workspace.Workspace) error {
	settings, err := json.Marshal(ws.Settings)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO workspaces (id, name, slug, settings, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err = r.db.ExecContext(ctx, query,
		ws.ID.String(),
		ws.Name,
		ws.Slug.String(),
		settings,
		ws.CreatedAt,
		ws.UpdatedAt,
	)
	return err
}

func (r *WorkspaceRepository) GetByID(ctx context.Context, id common.ID) (*workspace.Workspace, error) {
	query := `
		SELECT id, name, slug, settings, created_at, updated_at
		FROM workspaces
		WHERE id = $1
	`

	var ws workspace.Workspace
	var settingsJSON []byte

	err := r.db.QueryRowContext(ctx, query, id.String()).Scan(
		&ws.ID,
		&ws.Name,
		&ws.Slug,
		&settingsJSON,
		&ws.CreatedAt,
		&ws.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(settingsJSON, &ws.Settings); err != nil {
		return nil, err
	}

	return &ws, nil
}

func (r *WorkspaceRepository) GetBySlug(ctx context.Context, slug common.Slug) (*workspace.Workspace, error) {
	query := `
		SELECT id, name, slug, settings, created_at, updated_at
		FROM workspaces
		WHERE slug = $1
	`

	var ws workspace.Workspace
	var settingsJSON []byte

	err := r.db.QueryRowContext(ctx, query, slug.String()).Scan(
		&ws.ID,
		&ws.Name,
		&ws.Slug,
		&settingsJSON,
		&ws.CreatedAt,
		&ws.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(settingsJSON, &ws.Settings); err != nil {
		return nil, err
	}

	return &ws, nil
}

func (r *WorkspaceRepository) Update(ctx context.Context, ws *workspace.Workspace) error {
	settings, err := json.Marshal(ws.Settings)
	if err != nil {
		return err
	}

	query := `
		UPDATE workspaces
		SET name = $2, slug = $3, settings = $4, updated_at = $5
		WHERE id = $1
	`

	_, err = r.db.ExecContext(ctx, query,
		ws.ID.String(),
		ws.Name,
		ws.Slug.String(),
		settings,
		ws.UpdatedAt,
	)
	return err
}

func (r *WorkspaceRepository) Delete(ctx context.Context, id common.ID) error {
	query := `
		DELETE FROM workspaces
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, id.String())
	return err
}

func (r *WorkspaceRepository) List(ctx context.Context, limit, offset int) ([]*workspace.Workspace, error) {
	query := `
		SELECT id, name, slug, settings, created_at, updated_at
		FROM workspaces
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workspaces []*workspace.Workspace
	for rows.Next() {
		var ws workspace.Workspace
		var settingsJSON []byte

		if err := rows.Scan(
			&ws.ID,
			&ws.Name,
			&ws.Slug,
			&settingsJSON,
			&ws.CreatedAt,
			&ws.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if err := json.Unmarshal(settingsJSON, &ws.Settings); err != nil {
			return nil, err
		}

		workspaces = append(workspaces, &ws)
	}

	return workspaces, rows.Err()
}
