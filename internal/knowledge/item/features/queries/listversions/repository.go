package listversions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/lib/pq"

	itemdomain "documind.jordi.org/internal/knowledge/domain"
	shareddomain "documind.jordi.org/internal/shared/domain"
)

// Repository defines the persistence methods this use case needs.
type Repository interface {
	ListByItemID(ctx context.Context, itemID shareddomain.ID, limit, offset int) ([]*itemdomain.ItemVersion, error)
}

type postgresRepo struct {
	db *sql.DB
}

func newPostgresRepo(db *sql.DB) *postgresRepo {
	return &postgresRepo{db: db}
}

func (r *postgresRepo) ListByItemID(ctx context.Context, itemID shareddomain.ID, limit, offset int) ([]*itemdomain.ItemVersion, error) {
	query := `
		SELECT id, item_id, workspace_id, version,
			title, summary, body_md, custom_fields, tags, status,
			created_at, created_by
		FROM item_versions
		WHERE item_id = $1
		ORDER BY version DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, itemID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list item versions: %w", err)
	}
	defer rows.Close()

	var versions []*itemdomain.ItemVersion
	for rows.Next() {
		var itemVersion itemdomain.ItemVersion
		var customFieldsJSON []byte

		err := rows.Scan(
			&itemVersion.ID, &itemVersion.ItemID, &itemVersion.WorkspaceID, &itemVersion.Version,
			&itemVersion.Title, &itemVersion.Summary, &itemVersion.BodyMd, &customFieldsJSON,
			pq.Array(&itemVersion.Tags), &itemVersion.Status,
			&itemVersion.CreatedAt, &itemVersion.CreatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan item version: %w", err)
		}

		if err := json.Unmarshal(customFieldsJSON, &itemVersion.CustomFields); err != nil {
			return nil, fmt.Errorf("failed to unmarshal custom fields: %w", err)
		}

		versions = append(versions, &itemVersion)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating item versions: %w", err)
	}

	return versions, nil
}
