package get

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	shared "documind.jordi.org/internal/shared/domain"
	wsdomain "documind.jordi.org/internal/identity/workspace/domain"
)

type Repository interface {
	GetByID(ctx context.Context, id shared.ID) (*wsdomain.Workspace, error)
}

type postgresRepo struct{ db *sql.DB }

func newPostgresRepo(db *sql.DB) *postgresRepo { return &postgresRepo{db: db} }

func (r *postgresRepo) GetByID(ctx context.Context, id shared.ID) (*wsdomain.Workspace, error) {
	query := `
		SELECT id, name, slug, settings, created_at, updated_at
		FROM workspaces WHERE id = $1
	`
	var ws wsdomain.Workspace
	var settingsJSON []byte
	err := r.db.QueryRowContext(ctx, query, id.String()).Scan(
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
