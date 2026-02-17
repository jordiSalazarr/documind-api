package me

import (
	"context"
	"database/sql"
	"errors"
)

type Repository interface {
	GetByID(ctx context.Context, userID string) (*Result, error)
}

type postgresRepo struct {
	db *sql.DB
}

func newPostgresRepo(db *sql.DB) *postgresRepo {
	return &postgresRepo{db: db}
}

func (r *postgresRepo) GetByID(ctx context.Context, userID string) (*Result, error) {
	var result Result
	err := r.db.QueryRowContext(ctx,
		`SELECT id, email, name FROM users WHERE id = $1`,
		userID,
	).Scan(&result.ID, &result.Email, &result.Name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &result, nil
}
