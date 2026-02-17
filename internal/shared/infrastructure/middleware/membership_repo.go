package middleware

import (
	"context"
	"database/sql"
	"errors"
)

var ErrNotMember = errors.New("user is not a member of this workspace")

type MembershipChecker interface {
	CheckMembership(ctx context.Context, workspaceID, userID string) (role string, err error)
}

type postgresMembershipChecker struct {
	db *sql.DB
}

func NewMembershipChecker(db *sql.DB) MembershipChecker {
	return &postgresMembershipChecker{db: db}
}

func (r *postgresMembershipChecker) CheckMembership(ctx context.Context, workspaceID, userID string) (string, error) {
	var role string
	err := r.db.QueryRowContext(ctx,
		`SELECT role FROM workspace_members WHERE workspace_id = $1 AND user_id = $2`,
		workspaceID, userID,
	).Scan(&role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrNotMember
		}
		return "", err
	}
	return role, nil
}
