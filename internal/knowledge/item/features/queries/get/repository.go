package get

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
	GetByID(ctx context.Context, id shareddomain.ID) (*itemdomain.Item, error)
	GetLatestVersion(ctx context.Context, itemID shareddomain.ID) (*itemdomain.ItemVersion, error)
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
