package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/lib/pq"

	"documind.jordi.org/internal/domain/common"
	"documind.jordi.org/internal/domain/knowledge"
)

type ItemRepository struct {
	db *sql.DB
}

func NewItemRepository(db *sql.DB) *ItemRepository {
	return &ItemRepository{db: db}
}

// ===== Item Methods =====

func (r *ItemRepository) CreateItem(ctx context.Context, item *knowledge.Item) error {
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

func (r *ItemRepository) GetItemByID(ctx context.Context, id common.ID) (*knowledge.Item, error) {
	query := `
		SELECT id, workspace_id, project_id, service_id, item_type_id,
			latest_version, status, owner_user_id,
			created_at, updated_at, created_by, updated_by, deleted_at
		FROM items
		WHERE id = $1 AND deleted_at IS NULL
	`

	var item knowledge.Item
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

func (r *ItemRepository) ListItemsByProject(ctx context.Context, projectID common.ID) ([]*knowledge.Item, error) {
	query := `
		SELECT id, workspace_id, project_id, service_id, item_type_id,
			latest_version, status, owner_user_id,
			created_at, updated_at, created_by, updated_by, deleted_at
		FROM items
		WHERE project_id = $1 AND deleted_at IS NULL
		ORDER BY updated_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list items: %w", err)
	}
	defer rows.Close()

	var items []*knowledge.Item
	for rows.Next() {
		var item knowledge.Item
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

func (r *ItemRepository) ListItemsByService(ctx context.Context, serviceID common.ID) ([]*knowledge.Item, error) {
	query := `
		SELECT id, workspace_id, project_id, service_id, item_type_id,
			latest_version, status, owner_user_id,
			created_at, updated_at, created_by, updated_by, deleted_at
		FROM items
		WHERE service_id = $1 AND deleted_at IS NULL
		ORDER BY updated_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, serviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to list items by service: %w", err)
	}
	defer rows.Close()

	var items []*knowledge.Item
	for rows.Next() {
		var item knowledge.Item
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

func (r *ItemRepository) UpdateItem(ctx context.Context, item *knowledge.Item) error {
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

func (r *ItemRepository) DeleteItem(ctx context.Context, id common.ID) error {
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

// ===== ItemVersion Methods =====

func (r *ItemRepository) CreateItemVersion(ctx context.Context, version *knowledge.ItemVersion) error {
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

func (r *ItemRepository) GetItemVersion(ctx context.Context, itemID common.ID, version int) (*knowledge.ItemVersion, error) {
	query := `
		SELECT id, item_id, workspace_id, version,
			title, summary, body_md, custom_fields, tags, status,
			created_at, created_by
		FROM item_versions
		WHERE item_id = $1 AND version = $2
	`

	var itemVersion knowledge.ItemVersion
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

func (r *ItemRepository) GetLatestItemVersion(ctx context.Context, itemID common.ID) (*knowledge.ItemVersion, error) {
	query := `
		SELECT id, item_id, workspace_id, version,
			title, summary, body_md, custom_fields, tags, status,
			created_at, created_by
		FROM item_versions
		WHERE item_id = $1
		ORDER BY version DESC
		LIMIT 1
	`

	var itemVersion knowledge.ItemVersion
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

func (r *ItemRepository) GetItemVersionByID(ctx context.Context, versionID common.ID) (*knowledge.ItemVersion, error) {
	query := `
		SELECT id, item_id, workspace_id, version,
			title, summary, body_md, custom_fields, tags, status,
			created_at, created_by
		FROM item_versions
		WHERE id = $1
	`

	var itemVersion knowledge.ItemVersion
	var customFieldsJSON []byte

	err := r.db.QueryRowContext(ctx, query, versionID).Scan(
		&itemVersion.ID, &itemVersion.ItemID, &itemVersion.WorkspaceID, &itemVersion.Version,
		&itemVersion.Title, &itemVersion.Summary, &itemVersion.BodyMd, &customFieldsJSON,
		pq.Array(&itemVersion.Tags), &itemVersion.Status,
		&itemVersion.CreatedAt, &itemVersion.CreatedBy,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("item version not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get item version by ID: %w", err)
	}

	if err := json.Unmarshal(customFieldsJSON, &itemVersion.CustomFields); err != nil {
		return nil, fmt.Errorf("failed to unmarshal custom fields: %w", err)
	}

	return &itemVersion, nil
}

func (r *ItemRepository) ListItemVersions(ctx context.Context, itemID common.ID) ([]*knowledge.ItemVersion, error) {
	query := `
		SELECT id, item_id, workspace_id, version,
			title, summary, body_md, custom_fields, tags, status,
			created_at, created_by
		FROM item_versions
		WHERE item_id = $1
		ORDER BY version DESC
	`

	rows, err := r.db.QueryContext(ctx, query, itemID)
	if err != nil {
		return nil, fmt.Errorf("failed to list item versions: %w", err)
	}
	defer rows.Close()

	var versions []*knowledge.ItemVersion
	for rows.Next() {
		var itemVersion knowledge.ItemVersion
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

func (r *ItemRepository) UpdateItemVersionStatus(ctx context.Context, itemID common.ID, version int, status knowledge.ItemStatus) error {
	query := `
		UPDATE item_versions
		SET status = $3
		WHERE item_id = $1 AND version = $2
	`

	result, err := r.db.ExecContext(ctx, query, itemID, version, status)
	if err != nil {
		return fmt.Errorf("failed to update item version status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return errors.New("item version not found")
	}

	return nil
}

func (r *ItemRepository) ItemExists(ctx context.Context, id common.ID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM items WHERE id = $1 AND deleted_at IS NULL)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check item existence: %w", err)
	}

	return exists, nil
}

// SearchItemVersions performs full-text search across item versions
func (r *ItemRepository) SearchItemVersions(ctx context.Context, workspaceID common.ID, query string, limit, offset int) ([]*knowledge.ItemVersion, error) {
	searchQuery := `
		SELECT id, item_id, workspace_id, version,
			title, summary, body_md, custom_fields, tags, status,
			created_at, created_by,
			ts_rank(search_vector, websearch_to_tsquery('english', $2)) as rank
		FROM item_versions
		WHERE workspace_id = $1
			AND search_vector @@ websearch_to_tsquery('english', $2)
		ORDER BY rank DESC, created_at DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.db.QueryContext(ctx, searchQuery, workspaceID, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search item versions: %w", err)
	}
	defer rows.Close()

	var versions []*knowledge.ItemVersion
	for rows.Next() {
		var itemVersion knowledge.ItemVersion
		var customFieldsJSON []byte
		var rank float64

		err := rows.Scan(
			&itemVersion.ID, &itemVersion.ItemID, &itemVersion.WorkspaceID, &itemVersion.Version,
			&itemVersion.Title, &itemVersion.Summary, &itemVersion.BodyMd, &customFieldsJSON,
			pq.Array(&itemVersion.Tags), &itemVersion.Status,
			&itemVersion.CreatedAt, &itemVersion.CreatedBy,
			&rank,
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

// SearchItemVersionsByProject performs full-text search across item versions within a project
func (r *ItemRepository) SearchItemVersionsByProject(ctx context.Context, projectID common.ID, query string, limit, offset int) ([]*knowledge.ItemVersion, error) {
	searchQuery := `
		SELECT iv.id, iv.item_id, iv.workspace_id, iv.version,
			iv.title, iv.summary, iv.body_md, iv.custom_fields, iv.tags, iv.status,
			iv.created_at, iv.created_by,
			ts_rank(iv.search_vector, websearch_to_tsquery('english', $2)) as rank
		FROM item_versions iv
		JOIN items i ON iv.item_id = i.id
		WHERE i.project_id = $1
			AND i.deleted_at IS NULL
			AND iv.search_vector @@ websearch_to_tsquery('english', $2)
		ORDER BY rank DESC, iv.created_at DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.db.QueryContext(ctx, searchQuery, projectID, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search item versions by project: %w", err)
	}
	defer rows.Close()

	var versions []*knowledge.ItemVersion
	for rows.Next() {
		var itemVersion knowledge.ItemVersion
		var customFieldsJSON []byte
		var rank float64

		err := rows.Scan(
			&itemVersion.ID, &itemVersion.ItemID, &itemVersion.WorkspaceID, &itemVersion.Version,
			&itemVersion.Title, &itemVersion.Summary, &itemVersion.BodyMd, &customFieldsJSON,
			pq.Array(&itemVersion.Tags), &itemVersion.Status,
			&itemVersion.CreatedAt, &itemVersion.CreatedBy,
			&rank,
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

// UpdateItemVersionEmbedding updates the embedding vector for an item version
func (r *ItemRepository) UpdateItemVersionEmbedding(ctx context.Context, versionID common.ID, embedding []float32) error {
	query := `
		UPDATE item_versions
		SET embedding = $2
		WHERE id = $1
	`

	// Convert []float32 to string representation for pgvector
	embeddingStr := vectorToString(embedding)

	result, err := r.db.ExecContext(ctx, query, versionID, embeddingStr)
	if err != nil {
		return fmt.Errorf("failed to update embedding: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return errors.New("item version not found")
	}

	return nil
}

// SearchItemVersionsByEmbedding performs semantic search using vector similarity
func (r *ItemRepository) SearchItemVersionsByEmbedding(ctx context.Context, workspaceID common.ID, embedding []float32, limit, offset int) ([]*knowledge.ItemVersion, error) {
	embeddingStr := vectorToString(embedding)

	query := `
		SELECT id, item_id, workspace_id, version,
			title, summary, body_md, custom_fields, tags, status,
			created_at, created_by,
			1 - (embedding <=> $2::vector) as similarity
		FROM item_versions
		WHERE workspace_id = $1
			AND embedding IS NOT NULL
		ORDER BY embedding <=> $2::vector
		LIMIT $3 OFFSET $4
	`

	rows, err := r.db.QueryContext(ctx, query, workspaceID, embeddingStr, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search by embedding: %w", err)
	}
	defer rows.Close()

	var versions []*knowledge.ItemVersion
	for rows.Next() {
		var itemVersion knowledge.ItemVersion
		var customFieldsJSON []byte
		var similarity float64

		err := rows.Scan(
			&itemVersion.ID, &itemVersion.ItemID, &itemVersion.WorkspaceID, &itemVersion.Version,
			&itemVersion.Title, &itemVersion.Summary, &itemVersion.BodyMd, &customFieldsJSON,
			pq.Array(&itemVersion.Tags), &itemVersion.Status,
			&itemVersion.CreatedAt, &itemVersion.CreatedBy,
			&similarity,
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

// SearchItemVersionsByEmbeddingInProject performs semantic search within a project
func (r *ItemRepository) SearchItemVersionsByEmbeddingInProject(ctx context.Context, projectID common.ID, embedding []float32, limit, offset int) ([]*knowledge.ItemVersion, error) {
	embeddingStr := vectorToString(embedding)

	query := `
		SELECT iv.id, iv.item_id, iv.workspace_id, iv.version,
			iv.title, iv.summary, iv.body_md, iv.custom_fields, iv.tags, iv.status,
			iv.created_at, iv.created_by,
			1 - (iv.embedding <=> $2::vector) as similarity
		FROM item_versions iv
		JOIN items i ON iv.item_id = i.id
		WHERE i.project_id = $1
			AND i.deleted_at IS NULL
			AND iv.embedding IS NOT NULL
		ORDER BY iv.embedding <=> $2::vector
		LIMIT $3 OFFSET $4
	`

	rows, err := r.db.QueryContext(ctx, query, projectID, embeddingStr, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search by embedding in project: %w", err)
	}
	defer rows.Close()

	var versions []*knowledge.ItemVersion
	for rows.Next() {
		var itemVersion knowledge.ItemVersion
		var customFieldsJSON []byte
		var similarity float64

		err := rows.Scan(
			&itemVersion.ID, &itemVersion.ItemID, &itemVersion.WorkspaceID, &itemVersion.Version,
			&itemVersion.Title, &itemVersion.Summary, &itemVersion.BodyMd, &customFieldsJSON,
			pq.Array(&itemVersion.Tags), &itemVersion.Status,
			&itemVersion.CreatedAt, &itemVersion.CreatedBy,
			&similarity,
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

// vectorToString converts a float32 slice to pgvector string format
func vectorToString(vec []float32) string {
	if len(vec) == 0 {
		return "[]"
	}

	result := "["
	for i, v := range vec {
		if i > 0 {
			result += ","
		}
		result += fmt.Sprintf("%f", v)
	}
	result += "]"
	return result
}
