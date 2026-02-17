package delete

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	shareddomain "documind.jordi.org/internal/shared/domain"
)

// Repository defines the persistence methods this use case needs.
type Repository interface {
	Delete(ctx context.Context, id shareddomain.ID) error
}

type postgresRepo struct {
	db *sql.DB
}

func newPostgresRepo(db *sql.DB) *postgresRepo {
	return &postgresRepo{db: db}
}

func (r *postgresRepo) Delete(ctx context.Context, id shareddomain.ID) error {
	query := `
		UPDATE items
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete item: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return errors.New("item not found")
	}

	return nil
}
