package updatereport

import (
	"context"
	"database/sql"
)

// Repository defines persistence for updating staleness reports.
type Repository interface {
	UpdateStatus(ctx context.Context, id, workspaceID, status string) (bool, error)
}

type postgresRepo struct{ db *sql.DB }

func newPostgresRepo(db *sql.DB) *postgresRepo { return &postgresRepo{db: db} }

func (r *postgresRepo) UpdateStatus(ctx context.Context, id, workspaceID, status string) (bool, error) {
	query := `
		UPDATE staleness_reports SET status = $1
		WHERE id = $2 AND workspace_id = $3 AND deleted_at IS NULL
	`
	res, err := r.db.ExecContext(ctx, query, status, id, workspaceID)
	if err != nil {
		return false, err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return rows > 0, nil
}
