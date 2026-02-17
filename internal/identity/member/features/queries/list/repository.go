package list

import (
	"context"
	"database/sql"
)

type Repository interface {
	ListByWorkspace(ctx context.Context, workspaceID string) ([]MemberResult, error)
}

type postgresRepo struct {
	db *sql.DB
}

func newPostgresRepo(db *sql.DB) *postgresRepo {
	return &postgresRepo{db: db}
}

func (r *postgresRepo) ListByWorkspace(ctx context.Context, workspaceID string) ([]MemberResult, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, email, role, invited_at, joined_at
		 FROM workspace_members
		 WHERE workspace_id = $1
		 ORDER BY invited_at ASC`,
		workspaceID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []MemberResult
	for rows.Next() {
		var m MemberResult
		if err := rows.Scan(&m.ID, &m.UserID, &m.Email, &m.Role, &m.InvitedAt, &m.JoinedAt); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, rows.Err()
}
