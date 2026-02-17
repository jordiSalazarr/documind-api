package create

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/lib/pq"

	itemdomain "documind.jordi.org/internal/knowledge/domain"
)

// Repository defines the persistence methods this use case needs.
type Repository interface {
	CreateItem(ctx context.Context, item *itemdomain.Item) error
	UpdateItem(ctx context.Context, item *itemdomain.Item) error
	CreateItemVersion(ctx context.Context, version *itemdomain.ItemVersion) error
	ProjectExists(ctx context.Context, id string) (bool, error)
	ItemTypeExists(ctx context.Context, id string) (bool, error)
}

type postgresRepo struct {
	db *sql.DB
}

func newPostgresRepo(db *sql.DB) *postgresRepo {
	return &postgresRepo{db: db}
}

func (r *postgresRepo) CreateItem(ctx context.Context, item *itemdomain.Item) error {
	query := `
		INSERT INTO items (
			id, workspace_id, project_id, service_id, item_type_id,
			latest_version, status, owner_user_id,
			created_at, updated_at, created_by, updated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err := r.db.ExecContext(ctx, query,
		item.ID, item.WorkspaceID, item.ProjectID, item.ServiceID, item.ItemTypeID,
		item.LatestVersion, item.Status, item.OwnerUserID,
		item.CreatedAt, item.UpdatedAt, item.CreatedBy, item.UpdatedBy,
	)

	if err != nil {
		return fmt.Errorf("failed to create item: %w", err)
	}

	return nil
}

func (r *postgresRepo) UpdateItem(ctx context.Context, item *itemdomain.Item) error {
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

func (r *postgresRepo) CreateItemVersion(ctx context.Context, version *itemdomain.ItemVersion) error {
	customFieldsJSON, err := json.Marshal(version.CustomFields)
	if err != nil {
		return fmt.Errorf("failed to marshal custom fields: %w", err)
	}

	query := `
		INSERT INTO item_versions (
			id, item_id, workspace_id, version,
			title, summary, body_md, custom_fields, tags, status,
			created_at, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err = r.db.ExecContext(ctx, query,
		version.ID, version.ItemID, version.WorkspaceID, version.Version,
		version.Title, version.Summary, version.BodyMd, customFieldsJSON, pq.Array(version.Tags), version.Status,
		version.CreatedAt, version.CreatedBy,
	)

	if err != nil {
		return fmt.Errorf("failed to create item version: %w", err)
	}

	return nil
}

func (r *postgresRepo) ProjectExists(ctx context.Context, id string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM projects WHERE id = $1)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check project existence: %w", err)
	}
	return exists, nil
}

func (r *postgresRepo) ItemTypeExists(ctx context.Context, id string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM item_types WHERE id = $1)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check item type existence: %w", err)
	}
	return exists, nil
}
