package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"documind.jordi.org/internal/domain/common"
	"documind.jordi.org/internal/domain/knowledge"
)

type ItemTypeRepository struct {
	db *sql.DB
}

func NewItemTypeRepository(db *sql.DB) *ItemTypeRepository {
	return &ItemTypeRepository{db: db}
}

func (r *ItemTypeRepository) Create(ctx context.Context, itemType *knowledge.ItemType) error {
	schemaJSON, err := json.Marshal(itemType.SchemaJSON)
	if err != nil {
		return fmt.Errorf("failed to marshal schema: %w", err)
	}

	query := `
		INSERT INTO item_types (
			id, workspace_id, name, slug, schema_json, icon, color,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err = r.db.ExecContext(ctx, query,
		itemType.ID, itemType.WorkspaceID, itemType.Name, itemType.Slug,
		schemaJSON, itemType.Icon, itemType.Color,
		itemType.CreatedAt, itemType.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create item type: %w", err)
	}

	return nil
}

func (r *ItemTypeRepository) GetByID(ctx context.Context, id common.ID) (*knowledge.ItemType, error) {
	query := `
		SELECT id, workspace_id, name, slug, schema_json, icon, color,
			created_at, updated_at
		FROM item_types
		WHERE id = $1
	`

	var itemType knowledge.ItemType
	var schemaJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&itemType.ID, &itemType.WorkspaceID, &itemType.Name, &itemType.Slug,
		&schemaJSON, &itemType.Icon, &itemType.Color,
		&itemType.CreatedAt, &itemType.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("item type not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get item type: %w", err)
	}

	if err := json.Unmarshal(schemaJSON, &itemType.SchemaJSON); err != nil {
		return nil, fmt.Errorf("failed to unmarshal schema: %w", err)
	}

	return &itemType, nil
}

func (r *ItemTypeRepository) GetBySlug(ctx context.Context, workspaceID common.ID, slug common.Slug) (*knowledge.ItemType, error) {
	query := `
		SELECT id, workspace_id, name, slug, schema_json, icon, color,
			created_at, updated_at
		FROM item_types
		WHERE workspace_id = $1 AND slug = $2
	`

	var itemType knowledge.ItemType
	var schemaJSON []byte

	err := r.db.QueryRowContext(ctx, query, workspaceID, slug).Scan(
		&itemType.ID, &itemType.WorkspaceID, &itemType.Name, &itemType.Slug,
		&schemaJSON, &itemType.Icon, &itemType.Color,
		&itemType.CreatedAt, &itemType.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("item type not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get item type by slug: %w", err)
	}

	if err := json.Unmarshal(schemaJSON, &itemType.SchemaJSON); err != nil {
		return nil, fmt.Errorf("failed to unmarshal schema: %w", err)
	}

	return &itemType, nil
}

func (r *ItemTypeRepository) ListByWorkspace(ctx context.Context, workspaceID common.ID) ([]*knowledge.ItemType, error) {
	query := `
		SELECT id, workspace_id, name, slug, schema_json, icon, color,
			created_at, updated_at
		FROM item_types
		WHERE workspace_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to list item types: %w", err)
	}
	defer rows.Close()

	var itemTypes []*knowledge.ItemType
	for rows.Next() {
		var itemType knowledge.ItemType
		var schemaJSON []byte

		err := rows.Scan(
			&itemType.ID, &itemType.WorkspaceID, &itemType.Name, &itemType.Slug,
			&schemaJSON, &itemType.Icon, &itemType.Color,
			&itemType.CreatedAt, &itemType.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan item type: %w", err)
		}

		if err := json.Unmarshal(schemaJSON, &itemType.SchemaJSON); err != nil {
			return nil, fmt.Errorf("failed to unmarshal schema: %w", err)
		}

		itemTypes = append(itemTypes, &itemType)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating item types: %w", err)
	}

	return itemTypes, nil
}

func (r *ItemTypeRepository) Update(ctx context.Context, itemType *knowledge.ItemType) error {
	schemaJSON, err := json.Marshal(itemType.SchemaJSON)
	if err != nil {
		return fmt.Errorf("failed to marshal schema: %w", err)
	}

	query := `
		UPDATE item_types
		SET name = $2, schema_json = $3, icon = $4, color = $5, updated_at = $6
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		itemType.ID, itemType.Name, schemaJSON, itemType.Icon, itemType.Color, itemType.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update item type: %w", err)
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

func (r *ItemTypeRepository) Delete(ctx context.Context, id common.ID) error {
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

func (r *ItemTypeRepository) Exists(ctx context.Context, id common.ID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM item_types WHERE id = $1)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check item type existence: %w", err)
	}

	return exists, nil
}
