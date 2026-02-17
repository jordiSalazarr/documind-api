package list

import (
	"context"
	"database/sql"
)

// Repository defines read access for API keys.
type Repository interface {
	ListByWorkspace(ctx context.Context, workspaceID string) ([]*Result, error)
}

type postgresRepo struct{ db *sql.DB }

func newPostgresRepo(db *sql.DB) *postgresRepo { return &postgresRepo{db: db} }

func (r *postgresRepo) ListByWorkspace(ctx context.Context, workspaceID string) ([]*Result, error) {
	query := `
		SELECT id, name, key_prefix, created_by, expires_at, last_used_at, created_at
		FROM api_keys
		WHERE workspace_id = $1 AND revoked_at IS NULL
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*Result
	for rows.Next() {
		var r Result
		var expiresAt, lastUsedAt sql.NullTime
		if err := rows.Scan(&r.ID, &r.Name, &r.KeyPrefix, &r.CreatedBy, &expiresAt, &lastUsedAt, &r.CreatedAt); err != nil {
			return nil, err
		}
		if expiresAt.Valid {
			s := expiresAt.Time.Format("2006-01-02T15:04:05Z07:00")
			r.ExpiresAt = &s
		}
		if lastUsedAt.Valid {
			s := lastUsedAt.Time.Format("2006-01-02T15:04:05Z07:00")
			r.LastUsedAt = &s
		}
		results = append(results, &r)
	}
	return results, rows.Err()
}
