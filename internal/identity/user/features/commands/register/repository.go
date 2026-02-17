package register

import (
	"context"
	"database/sql"
	"errors"

	userdomain "documind.jordi.org/internal/identity/user/domain"
	"github.com/lib/pq"
)

type Repository interface {
	Insert(ctx context.Context, user *userdomain.User) error
}

type postgresRepo struct {
	db *sql.DB
}

func newPostgresRepo(db *sql.DB) *postgresRepo {
	return &postgresRepo{db: db}
}

func (r *postgresRepo) Insert(ctx context.Context, user *userdomain.User) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO users (id, email, password_hash, name, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		string(user.ID),
		string(user.Email),
		user.PasswordHash,
		user.Name,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return ErrEmailTaken
		}
		return err
	}
	return nil
}
