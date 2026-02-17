package list

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"documind.jordi.org/internal/knowledge/domain"
	shareddomain "documind.jordi.org/internal/shared/domain"
)

type Repository interface {
	ListByWorkspace(ctx context.Context, workspaceID string, limit, offset int) ([]*domain.ItemType, error)
}

type postgresRepo struct{ db *sql.DB }

func newPostgresRepo(db *sql.DB) *postgresRepo { return &postgresRepo{db: db} }

func (r *postgresRepo) ListByWorkspace(ctx context.Context, workspaceID string, limit, offset int) ([]*domain.ItemType, error) {
	query := `
		SELECT id, workspace_id, name, slug, description, icon, schema_json, created_at, updated_at
		FROM item_types WHERE workspace_id = $1 ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, workspaceID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list item types: %w", err)
	}
	defer rows.Close()
	var itemTypes []*domain.ItemType
	for rows.Next() {
		var it domain.ItemType
		var id, wsID, name, slugStr string
		var description, icon *string
		var fieldsJSON []byte
		if err := rows.Scan(&id, &wsID, &name, &slugStr, &description, &icon, &fieldsJSON, &it.CreatedAt, &it.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan item type: %w", err)
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
		itemTypes = append(itemTypes, &it)
	}
	return itemTypes, rows.Err()
}
