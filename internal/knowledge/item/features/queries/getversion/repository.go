package getversion

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
	GetVersion(ctx context.Context, itemID shareddomain.ID, version int) (*itemdomain.ItemVersion, error)
}

type postgresRepo struct {
	db *sql.DB
}

func newPostgresRepo(db *sql.DB) *postgresRepo {
	return &postgresRepo{db: db}
}

func (r *postgresRepo) GetVersion(ctx context.Context, itemID shareddomain.ID, version int) (*itemdomain.ItemVersion, error) {
	query := `
		SELECT id, item_id, workspace_id, version,
			title, summary, body_md, custom_fields, tags, status,
			created_at, created_by
		FROM item_versions
		WHERE item_id = $1 AND version = $2
	`

	var itemVersion itemdomain.ItemVersion
	var customFieldsJSON []byte

	err := r.db.QueryRowContext(ctx, query, itemID, version).Scan(
		&itemVersion.ID, &itemVersion.ItemID, &itemVersion.WorkspaceID, &itemVersion.Version,
		&itemVersion.Title, &itemVersion.Summary, &itemVersion.BodyMd, &customFieldsJSON,
		pq.Array(&itemVersion.Tags), &itemVersion.Status,
		&itemVersion.CreatedAt, &itemVersion.CreatedBy,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("item version not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get item version: %w", err)
	}

	if err := json.Unmarshal(customFieldsJSON, &itemVersion.CustomFields); err != nil {
		return nil, fmt.Errorf("failed to unmarshal custom fields: %w", err)
	}

	return &itemVersion, nil
}
