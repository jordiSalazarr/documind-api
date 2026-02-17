package update

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	itemdomain "documind.jordi.org/internal/knowledge/domain"
	shareddomain "documind.jordi.org/internal/shared/domain"
)

// Repository defines the persistence methods this use case needs.
type Repository interface {
	GetByID(ctx context.Context, id shareddomain.ID) (*itemdomain.Item, error)
	Update(ctx context.Context, item *itemdomain.Item) error
}

type postgresRepo struct {
	db *sql.DB
}

func newPostgresRepo(db *sql.DB) *postgresRepo {
	return &postgresRepo{db: db}
}

func (r *postgresRepo) GetByID(ctx context.Context, id shareddomain.ID) (*itemdomain.Item, error) {
	query := `
		SELECT id, workspace_id, project_id, service_id, item_type_id,
			latest_version, status, owner_user_id,
			created_at, updated_at, created_by, updated_by, deleted_at
		FROM items
		WHERE id = $1 AND deleted_at IS NULL
	`

	var item itemdomain.Item
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&item.ID, &item.WorkspaceID, &item.ProjectID, &item.ServiceID, &item.ItemTypeID,
		&item.LatestVersion, &item.Status, &item.OwnerUserID,
		&item.CreatedAt, &item.UpdatedAt, &item.CreatedBy, &item.UpdatedBy, &item.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("item not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get item: %w", err)
	}

	return &item, nil
}

func (r *postgresRepo) Update(ctx context.Context, item *itemdomain.Item) error {
	query := `
		UPDATE items
		SET project_id = $2, service_id = $3, item_type_id = $4,
			latest_version = $5, status = $6, owner_user_id = $7,
			updated_at = $8, updated_by = $9
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query,
		item.ID, item.ProjectID, item.ServiceID, item.ItemTypeID,
		item.LatestVersion, item.Status, item.OwnerUserID,
		item.UpdatedAt, item.UpdatedBy,
	)

	if err != nil {
		return fmt.Errorf("failed to update item: %w", err)
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
