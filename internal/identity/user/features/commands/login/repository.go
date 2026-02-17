package login

import (
	"context"
	"database/sql"
	"errors"

	userdomain "documind.jordi.org/internal/identity/user/domain"
	shared "documind.jordi.org/internal/shared/domain"
)

type Repository interface {
	GetByEmail(ctx context.Context, email string) (*userdomain.User, error)
}

type postgresRepo struct {
	db *sql.DB
}

func newPostgresRepo(db *sql.DB) *postgresRepo {
	return &postgresRepo{db: db}
}

func (r *postgresRepo) GetByEmail(ctx context.Context, email string) (*userdomain.User, error) {
	user := &userdomain.User{}
	var id, emailStr string

	err := r.db.QueryRowContext(ctx,
		`SELECT id, email, password_hash, name, created_at, updated_at
		 FROM users WHERE email = $1`,
		email,
	).Scan(&id, &emailStr, &user.PasswordHash, &user.Name, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	user.ID = shared.ID(id)
	user.Email = shared.Email(emailStr)
	return user, nil
}
