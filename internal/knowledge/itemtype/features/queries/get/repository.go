package get

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"documind.jordi.org/internal/knowledge/domain"
	shareddomain "documind.jordi.org/internal/shared/domain"
)

type Repository interface {
	GetByID(ctx context.Context, id string) (*domain.ItemType, error)
}

type postgresRepo struct{ db *sql.DB }

func newPostgresRepo(db *sql.DB) *postgresRepo { return &postgresRepo{db: db} }

func (r *postgresRepo) GetByID(ctx context.Context, id string) (*domain.ItemType, error) {
	query := `
		SELECT id, workspace_id, name, slug, description, icon, schema_json, created_at, updated_at
		FROM item_types WHERE id = $1
	`
	var it domain.ItemType
	var idStr, wsID, name, slugStr string
	var description, icon *string
	var fieldsJSON []byte
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&idStr, &wsID, &name, &slugStr, &description, &icon, &fieldsJSON,
		&it.CreatedAt, &it.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, errors.New("item type not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get item type: %w", err)
	}
	it.ID = shareddomain.ID(idStr)
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
