package create

import (
	"context"
	"database/sql"
)

// Repository defines persistence for API key creation.
type Repository interface {
	Insert(ctx context.Context, id, workspaceID, name, keyHash, keyPrefix, createdBy string) error
}

type postgresRepo struct{ db *sql.DB }

func newPostgresRepo(db *sql.DB) *postgresRepo { return &postgresRepo{db: db} }

func (r *postgresRepo) Insert(ctx context.Context, id, workspaceID, name, keyHash, keyPrefix, createdBy string) error {
	query := `
		INSERT INTO api_keys (id, workspace_id, name, key_hash, key_prefix, created_by)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.ExecContext(ctx, query, id, workspaceID, name, keyHash, keyPrefix, createdBy)
	return err
}
