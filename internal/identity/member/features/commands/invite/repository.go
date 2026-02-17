package invite

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"
)

type Repository interface {
	FindUserByEmail(ctx context.Context, email string) (userID string, err error)
	InsertMember(ctx context.Context, workspaceID, userID, email, role, invitedBy string) (memberID string, err error)
}

type postgresRepo struct {
	db *sql.DB
}

func newPostgresRepo(db *sql.DB) *postgresRepo {
	return &postgresRepo{db: db}
}

func (r *postgresRepo) FindUserByEmail(ctx context.Context, email string) (string, error) {
	var userID string
	err := r.db.QueryRowContext(ctx,
		`SELECT id FROM users WHERE email = $1`, email,
	).Scan(&userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrUserNotFound
		}
		return "", err
	}
	return userID, nil
}

func (r *postgresRepo) InsertMember(ctx context.Context, workspaceID, userID, email, role, invitedBy string) (string, error) {
	var memberID string
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO workspace_members (workspace_id, user_id, email, role, invited_by, joined_at)
		 VALUES ($1, $2, $3, $4, $5, NOW())
		 RETURNING id`,
		workspaceID, userID, email, role, invitedBy,
	).Scan(&memberID)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return "", ErrAlreadyMember
		}
		return "", err
	}
	return memberID, nil
}
