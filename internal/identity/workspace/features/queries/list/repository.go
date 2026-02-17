package list

import (
	"context"
	"database/sql"
	"encoding/json"

	wsdomain "documind.jordi.org/internal/identity/workspace/domain"
)

type Repository interface {
	ListByUser(ctx context.Context, userID string, limit, offset int) ([]*wsdomain.Workspace, error)
}

type postgresRepo struct{ db *sql.DB }

func newPostgresRepo(db *sql.DB) *postgresRepo { return &postgresRepo{db: db} }

func (r *postgresRepo) ListByUser(ctx context.Context, userID string, limit, offset int) ([]*wsdomain.Workspace, error) {
	query := `
		SELECT w.id, w.name, w.slug, w.settings, w.created_at, w.updated_at
		FROM workspaces w
		INNER JOIN workspace_members wm ON wm.workspace_id = w.id
		WHERE wm.user_id = $1
		ORDER BY w.created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workspaces []*wsdomain.Workspace
	for rows.Next() {
		var ws wsdomain.Workspace
		var settingsJSON []byte
		if err := rows.Scan(
			&ws.ID, &ws.Name, &ws.Slug, &settingsJSON, &ws.CreatedAt, &ws.UpdatedAt,
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
