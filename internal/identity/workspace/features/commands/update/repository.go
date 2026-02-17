package update

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
	Update(ctx context.Context, ws *wsdomain.Workspace) error
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

func (r *postgresRepo) Update(ctx context.Context, ws *wsdomain.Workspace) error {
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
		ws.ID.String(), ws.Name, ws.Slug.String(), settings, ws.UpdatedAt,
	)
	return err
}
