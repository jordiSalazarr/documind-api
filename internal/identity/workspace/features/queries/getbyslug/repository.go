package getbyslug

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	shared "documind.jordi.org/internal/shared/domain"
	wsdomain "documind.jordi.org/internal/identity/workspace/domain"
)

type Repository interface {
	GetBySlug(ctx context.Context, slug shared.Slug) (*wsdomain.Workspace, error)
}

type postgresRepo struct{ db *sql.DB }

func newPostgresRepo(db *sql.DB) *postgresRepo { return &postgresRepo{db: db} }

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
