package list

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/lib/pq"

	itemdomain "documind.jordi.org/internal/knowledge/domain"
	shareddomain "documind.jordi.org/internal/shared/domain"
)

// Repository defines the persistence methods this use case needs.
type Repository interface {
	ListByProject(ctx context.Context, projectID shareddomain.ID, limit, offset int) ([]*itemdomain.Item, error)
	ListByService(ctx context.Context, serviceID shareddomain.ID, limit, offset int) ([]*itemdomain.Item, error)
	CountByProject(ctx context.Context, projectID shareddomain.ID) (int, error)
	CountByService(ctx context.Context, serviceID shareddomain.ID) (int, error)
	GetLatestVersion(ctx context.Context, itemID shareddomain.ID) (*itemdomain.ItemVersion, error)
}

type postgresRepo struct {
	db *sql.DB
}

func newPostgresRepo(db *sql.DB) *postgresRepo {
	return &postgresRepo{db: db}
}

func (r *postgresRepo) ListByProject(ctx context.Context, projectID shareddomain.ID, limit, offset int) ([]*itemdomain.Item, error) {
	query := `
		SELECT id, workspace_id, project_id, service_id, item_type_id,
			latest_version, status, owner_user_id,
			created_at, updated_at, created_by, updated_by, deleted_at
		FROM items
		WHERE project_id = $1 AND deleted_at IS NULL
		ORDER BY updated_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, projectID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list items: %w", err)
	}
	defer rows.Close()

	var items []*itemdomain.Item
	for rows.Next() {
		var item itemdomain.Item
		err := rows.Scan(
			&item.ID, &item.WorkspaceID, &item.ProjectID, &item.ServiceID, &item.ItemTypeID,
			&item.LatestVersion, &item.Status, &item.OwnerUserID,
			&item.CreatedAt, &item.UpdatedAt, &item.CreatedBy, &item.UpdatedBy, &item.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan item: %w", err)
		}
		items = append(items, &item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating items: %w", err)
	}

	return items, nil
}

func (r *postgresRepo) ListByService(ctx context.Context, serviceID shareddomain.ID, limit, offset int) ([]*itemdomain.Item, error) {
	query := `
		SELECT id, workspace_id, project_id, service_id, item_type_id,
			latest_version, status, owner_user_id,
			created_at, updated_at, created_by, updated_by, deleted_at
		FROM items
		WHERE service_id = $1 AND deleted_at IS NULL
		ORDER BY updated_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, serviceID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list items by service: %w", err)
	}
	defer rows.Close()

	var items []*itemdomain.Item
	for rows.Next() {
		var item itemdomain.Item
		err := rows.Scan(
			&item.ID, &item.WorkspaceID, &item.ProjectID, &item.ServiceID, &item.ItemTypeID,
			&item.LatestVersion, &item.Status, &item.OwnerUserID,
			&item.CreatedAt, &item.UpdatedAt, &item.CreatedBy, &item.UpdatedBy, &item.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan item: %w", err)
		}
		items = append(items, &item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating items: %w", err)
	}

	return items, nil
}

func (r *postgresRepo) CountByProject(ctx context.Context, projectID shareddomain.ID) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM items WHERE project_id = $1 AND deleted_at IS NULL",
		projectID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count items by project: %w", err)
	}
	return count, nil
}

func (r *postgresRepo) CountByService(ctx context.Context, serviceID shareddomain.ID) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM items WHERE service_id = $1 AND deleted_at IS NULL",
		serviceID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count items by service: %w", err)
	}
	return count, nil
}

func (r *postgresRepo) GetLatestVersion(ctx context.Context, itemID shareddomain.ID) (*itemdomain.ItemVersion, error) {
	query := `
		SELECT id, item_id, workspace_id, version,
			title, summary, body_md, custom_fields, tags, status,
			created_at, created_by
		FROM item_versions
		WHERE item_id = $1
		ORDER BY version DESC
		LIMIT 1
	`

	var itemVersion itemdomain.ItemVersion
	var customFieldsJSON []byte

	err := r.db.QueryRowContext(ctx, query, itemID).Scan(
		&itemVersion.ID, &itemVersion.ItemID, &itemVersion.WorkspaceID, &itemVersion.Version,
		&itemVersion.Title, &itemVersion.Summary, &itemVersion.BodyMd, &customFieldsJSON,
		pq.Array(&itemVersion.Tags), &itemVersion.Status,
		&itemVersion.CreatedAt, &itemVersion.CreatedBy,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("item version not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get latest item version: %w", err)
	}

	if err := json.Unmarshal(customFieldsJSON, &itemVersion.CustomFields); err != nil {
		return nil, fmt.Errorf("failed to unmarshal custom fields: %w", err)
	}

	return &itemVersion, nil
}
