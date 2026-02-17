package create

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	shared "documind.jordi.org/internal/shared/domain"
	wsdomain "documind.jordi.org/internal/identity/workspace/domain"
)

type Repository interface {
	Insert(ctx context.Context, ws *wsdomain.Workspace) error
	GetBySlug(ctx context.Context, slug shared.Slug) (*wsdomain.Workspace, error)
	InsertMember(ctx context.Context, workspaceID, userID, email, role string) error
}

type postgresRepo struct{ db *sql.DB }

func newPostgresRepo(db *sql.DB) *postgresRepo { return &postgresRepo{db: db} }

func (r *postgresRepo) Insert(ctx context.Context, ws *wsdomain.Workspace) error {
	settings, err := json.Marshal(ws.Settings)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO workspaces (id, name, slug, settings, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err = r.db.ExecContext(ctx, query,
		ws.ID.String(), ws.Name, ws.Slug.String(), settings, ws.CreatedAt, ws.UpdatedAt,
	)
	return err
}

func (r *postgresRepo) InsertMember(ctx context.Context, workspaceID, userID, email, role string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO workspace_members (workspace_id, user_id, email, role, joined_at)
		 VALUES ($1, $2, $3, $4, NOW())`,
		workspaceID, userID, email, role,
	)
	return err
}

func (r *postgresRepo) GetBySlug(ctx context.Context, slug shared.Slug) (*wsdomain.Workspace, error) {
	query := `
		SELECT id, name, slug, settings, created_at, updated_at
		FROM workspaces WHERE slug = $1
	`
	var ws wsdomain.Workspace
	var settingsJSON []byte
	err := r.db.QueryRowContext(ctx, query, slug.String()).Scan(
		&ws.ID, &ws.Name, &ws.Slug, &settingsJSON, &ws.CreatedAt, &ws.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(settingsJSON, &ws.Settings); err != nil {
		return nil, err
	}
	return &ws, nil
}
