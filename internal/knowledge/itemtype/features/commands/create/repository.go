package create

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"documind.jordi.org/internal/knowledge/domain"
	shareddomain "documind.jordi.org/internal/shared/domain"
)

type Repository interface {
	Insert(ctx context.Context, itemType *domain.ItemType) error
	GetBySlug(ctx context.Context, workspaceID string, slug string) (*domain.ItemType, error)
}

type postgresRepo struct{ db *sql.DB }

func newPostgresRepo(db *sql.DB) *postgresRepo { return &postgresRepo{db: db} }

func (r *postgresRepo) Insert(ctx context.Context, itemType *domain.ItemType) error {
	fieldsJSON, err := json.Marshal(itemType.Fields)
	if err != nil {
		return fmt.Errorf("failed to marshal fields: %w", err)
	}
	query := `
		INSERT INTO item_types (id, workspace_id, name, slug, description, icon, schema_json, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err = r.db.ExecContext(ctx, query,
		itemType.ID, itemType.WorkspaceID, itemType.Name, itemType.Slug,
		itemType.Description, itemType.Icon, fieldsJSON,
		itemType.CreatedAt, itemType.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create item type: %w", err)
	}
	return nil
}

func (r *postgresRepo) GetBySlug(ctx context.Context, workspaceID string, slug string) (*domain.ItemType, error) {
	query := `
		SELECT id, workspace_id, name, slug, description, icon, schema_json, created_at, updated_at
		FROM item_types WHERE workspace_id = $1 AND slug = $2
	`
	var it domain.ItemType
	var id, wsID, name, slugStr string
	var description, icon *string
	var fieldsJSON []byte
	err := r.db.QueryRowContext(ctx, query, workspaceID, slug).Scan(
		&id, &wsID, &name, &slugStr, &description, &icon, &fieldsJSON,
		&it.CreatedAt, &it.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get item type: %w", err)
	}
	it.ID = shareddomain.ID(id)
	it.WorkspaceID = shareddomain.ID(wsID)
	it.Name = name
	it.Slug = shareddomain.Slug(slugStr)
	it.Description = description
	it.Icon = icon
	if fieldsJSON != nil {
		if err := json.Unmarshal(fieldsJSON, &it.Fields); err != nil {
			return nil, fmt.Errorf("failed to unmarshal fields: %w", err)
		}
	}
	return &it, nil
}
