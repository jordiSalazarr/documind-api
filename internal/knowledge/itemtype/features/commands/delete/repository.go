package delete

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type Repository interface {
	Exists(ctx context.Context, id string) (bool, error)
	Delete(ctx context.Context, id string) error
}

type postgresRepo struct{ db *sql.DB }

func newPostgresRepo(db *sql.DB) *postgresRepo { return &postgresRepo{db: db} }

func (r *postgresRepo) Exists(ctx context.Context, id string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM item_types WHERE id = $1)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check item type existence: %w", err)
	}
	return exists, nil
}

func (r *postgresRepo) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM item_types WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete item type: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return errors.New("item type not found")
	}
	return nil
}
