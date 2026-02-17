package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	shared "documind.jordi.org/internal/shared/domain"
	reldomain "documind.jordi.org/internal/knowledge/domain"
)

// RelationTypeRepo provides shared relation type persistence
type RelationTypeRepo struct{ DB *sql.DB }

func (r *RelationTypeRepo) GetByID(ctx context.Context, id shared.ID) (*reldomain.RelationType, error) {
	query := `
		SELECT id, workspace_id, name, slug, is_directional, created_at, updated_at, deleted_at
		FROM relation_types WHERE id = $1 AND deleted_at IS NULL
	`
	var rt reldomain.RelationType
	err := r.DB.QueryRowContext(ctx, query, id.String()).Scan(
		&rt.ID, &rt.WorkspaceID, &rt.Name, &rt.Slug, &rt.IsDirectional,
		&rt.CreatedAt, &rt.UpdatedAt, &rt.DeletedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get relation type: %w", err)
	}
	return &rt, nil
}

func (r *RelationTypeRepo) ListByWorkspace(ctx context.Context, workspaceID shared.ID, limit, offset int) ([]*reldomain.RelationType, error) {
	query := `
		SELECT id, workspace_id, name, slug, is_directional, created_at, updated_at, deleted_at
		FROM relation_types WHERE workspace_id = $1 AND deleted_at IS NULL ORDER BY name ASC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.DB.QueryContext(ctx, query, workspaceID.String(), limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list relation types: %w", err)
	}
	defer rows.Close()
	var types []*reldomain.RelationType
	for rows.Next() {
		var rt reldomain.RelationType
		if err := rows.Scan(&rt.ID, &rt.WorkspaceID, &rt.Name, &rt.Slug, &rt.IsDirectional, &rt.CreatedAt, &rt.UpdatedAt, &rt.DeletedAt); err != nil {
			return nil, fmt.Errorf("failed to scan relation type: %w", err)
		}
		types = append(types, &rt)
	}
	return types, rows.Err()
}

func (r *RelationTypeRepo) Exists(ctx context.Context, workspaceID shared.ID, slug shared.Slug) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM relation_types WHERE workspace_id = $1 AND slug = $2 AND deleted_at IS NULL)`
	var exists bool
	err := r.DB.QueryRowContext(ctx, query, workspaceID.String(), slug.String()).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check relation type existence: %w", err)
	}
	return exists, nil
}

func (r *RelationTypeRepo) Create(ctx context.Context, rt *reldomain.RelationType) error {
	query := `INSERT INTO relation_types (id, workspace_id, name, slug, is_directional, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.DB.ExecContext(ctx, query,
		rt.ID.String(), rt.WorkspaceID.String(), rt.Name, rt.Slug.String(), rt.IsDirectional, rt.CreatedAt, rt.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create relation type: %w", err)
	}
	return nil
}

func (r *RelationTypeRepo) Update(ctx context.Context, rt *reldomain.RelationType) error {
	query := `UPDATE relation_types SET name = $2, slug = $3, is_directional = $4, updated_at = $5 WHERE id = $1 AND deleted_at IS NULL`
	result, err := r.DB.ExecContext(ctx, query, rt.ID.String(), rt.Name, rt.Slug.String(), rt.IsDirectional, rt.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to update relation type: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return errors.New("relation type not found")
	}
	return nil
}

func (r *RelationTypeRepo) Delete(ctx context.Context, id shared.ID) error {
	query := `UPDATE relation_types SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	result, err := r.DB.ExecContext(ctx, query, id.String())
	if err != nil {
		return fmt.Errorf("failed to delete relation type: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return errors.New("relation type not found")
	}
	return nil
}

// RelationRepo provides shared relation persistence
type RelationRepo struct{ DB *sql.DB }

func (r *RelationRepo) GetByID(ctx context.Context, id shared.ID) (*reldomain.Relation, error) {
	query := `
		SELECT r.id, r.workspace_id, r.from_item_id, r.to_item_id, r.relation_type_id, r.created_at, r.created_by, r.deleted_at,
			rt.id, rt.workspace_id, rt.name, rt.slug, rt.is_directional, rt.created_at, rt.updated_at
		FROM item_relations r JOIN relation_types rt ON r.relation_type_id = rt.id
		WHERE r.id = $1 AND r.deleted_at IS NULL
	`
	var rel reldomain.Relation
	var rt reldomain.RelationType
	err := r.DB.QueryRowContext(ctx, query, id.String()).Scan(
		&rel.ID, &rel.WorkspaceID, &rel.FromItemID, &rel.ToItemID, &rel.RelationTypeID, &rel.CreatedAt, &rel.CreatedBy, &rel.DeletedAt,
		&rt.ID, &rt.WorkspaceID, &rt.Name, &rt.Slug, &rt.IsDirectional, &rt.CreatedAt, &rt.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get relation: %w", err)
	}
	rel.RelationType = &rt
	return &rel, nil
}

func (r *RelationRepo) ListByItem(ctx context.Context, itemID shared.ID, limit, offset int) ([]*reldomain.Relation, error) {
	query := `
		SELECT r.id, r.workspace_id, r.from_item_id, r.to_item_id, r.relation_type_id, r.created_at, r.created_by, r.deleted_at,
			rt.id, rt.workspace_id, rt.name, rt.slug, rt.is_directional, rt.created_at, rt.updated_at
		FROM item_relations r JOIN relation_types rt ON r.relation_type_id = rt.id
		WHERE (r.from_item_id = $1 OR r.to_item_id = $1) AND r.deleted_at IS NULL ORDER BY r.created_at DESC
		LIMIT $2 OFFSET $3
	`
	return r.queryRelations(ctx, query, itemID, limit, offset)
}

func (r *RelationRepo) Exists(ctx context.Context, fromItemID, toItemID, relationTypeID shared.ID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM item_relations WHERE from_item_id = $1 AND to_item_id = $2 AND relation_type_id = $3 AND deleted_at IS NULL)`
	var exists bool
	err := r.DB.QueryRowContext(ctx, query, fromItemID.String(), toItemID.String(), relationTypeID.String()).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check relation existence: %w", err)
	}
	return exists, nil
}

func (r *RelationRepo) Create(ctx context.Context, relation *reldomain.Relation) error {
	query := `INSERT INTO item_relations (id, workspace_id, from_item_id, to_item_id, relation_type_id, created_at, created_by) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.DB.ExecContext(ctx, query,
		relation.ID.String(), relation.WorkspaceID.String(), relation.FromItemID.String(), relation.ToItemID.String(),
		relation.RelationTypeID.String(), relation.CreatedAt, relation.CreatedBy.String(),
	)
	if err != nil {
		return fmt.Errorf("failed to create relation: %w", err)
	}
	return nil
}

func (r *RelationRepo) Delete(ctx context.Context, id shared.ID) error {
	query := `UPDATE item_relations SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	result, err := r.DB.ExecContext(ctx, query, id.String())
	if err != nil {
		return fmt.Errorf("failed to delete relation: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return errors.New("relation not found")
	}
	return nil
}

func (r *RelationRepo) queryRelations(ctx context.Context, query string, arg shared.ID, extraArgs ...interface{}) ([]*reldomain.Relation, error) {
	args := []interface{}{arg.String()}
	args = append(args, extraArgs...)
	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query relations: %w", err)
	}
	defer rows.Close()
	var relations []*reldomain.Relation
	for rows.Next() {
		var rel reldomain.Relation
		var rt reldomain.RelationType
		if err := rows.Scan(
			&rel.ID, &rel.WorkspaceID, &rel.FromItemID, &rel.ToItemID, &rel.RelationTypeID, &rel.CreatedAt, &rel.CreatedBy, &rel.DeletedAt,
			&rt.ID, &rt.WorkspaceID, &rt.Name, &rt.Slug, &rt.IsDirectional, &rt.CreatedAt, &rt.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan relation: %w", err)
		}
		rel.RelationType = &rt
		relations = append(relations, &rel)
	}
	return relations, rows.Err()
}
