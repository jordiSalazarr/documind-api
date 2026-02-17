package remove

import (
	"context"
	"database/sql"
	"errors"
)

type Repository interface {
	GetMemberRole(ctx context.Context, workspaceID, userID string) (string, error)
	CountAdmins(ctx context.Context, workspaceID string) (int, error)
	Delete(ctx context.Context, workspaceID, userID string) error
}

type postgresRepo struct {
	db *sql.DB
}

func newPostgresRepo(db *sql.DB) *postgresRepo {
	return &postgresRepo{db: db}
}

func (r *postgresRepo) GetMemberRole(ctx context.Context, workspaceID, userID string) (string, error) {
	var role string
	err := r.db.QueryRowContext(ctx,
		`SELECT role FROM workspace_members WHERE workspace_id = $1 AND user_id = $2`,
		workspaceID, userID,
	).Scan(&role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrMemberNotFound
		}
		return "", err
	}
	return role, nil
}

func (r *postgresRepo) CountAdmins(ctx context.Context, workspaceID string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM workspace_members WHERE workspace_id = $1 AND role = 'admin'`,
		workspaceID,
	).Scan(&count)
	return count, err
}

func (r *postgresRepo) Delete(ctx context.Context, workspaceID, userID string) error {
	result, err := r.db.ExecContext(ctx,
		`DELETE FROM workspace_members WHERE workspace_id = $1 AND user_id = $2`,
		workspaceID, userID,
	)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrMemberNotFound
	}
	return nil
}
