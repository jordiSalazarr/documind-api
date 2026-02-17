package revoke

import (
	"context"
	"database/sql"
)

// Repository defines persistence for API key revocation.
type Repository interface {
	Revoke(ctx context.Context, id, workspaceID string) (bool, error)
}

type postgresRepo struct{ db *sql.DB }

func newPostgresRepo(db *sql.DB) *postgresRepo { return &postgresRepo{db: db} }

func (r *postgresRepo) Revoke(ctx context.Context, id, workspaceID string) (bool, error) {
	query := `
		UPDATE api_keys SET revoked_at = NOW()
		WHERE id = $1 AND workspace_id = $2 AND revoked_at IS NULL
	`
	res, err := r.db.ExecContext(ctx, query, id, workspaceID)
	if err != nil {
		return false, err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return rows > 0, nil
}
