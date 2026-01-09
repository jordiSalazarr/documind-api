package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"documind.jordi.org/internal/domain/common"
	"documind.jordi.org/internal/domain/knowledge"
)

type RelationTypeRepository struct {
	db *sql.DB
}

func NewRelationTypeRepository(db *sql.DB) *RelationTypeRepository {
	return &RelationTypeRepository{db: db}
}

func (r *RelationTypeRepository) Create(ctx context.Context, relationType *knowledge.RelationType) error {
	query := `
		INSERT INTO relation_types (
			id, workspace_id, name, slug, is_directional,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.ExecContext(ctx, query,
		relationType.ID, relationType.WorkspaceID, relationType.Name,
		relationType.Slug, relationType.IsDirectional,
		relationType.CreatedAt, relationType.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create relation type: %w", err)
	}

	return nil
}

func (r *RelationTypeRepository) GetByID(ctx context.Context, id common.ID) (*knowledge.RelationType, error) {
	query := `
		SELECT id, workspace_id, name, slug, is_directional,
			created_at, updated_at, deleted_at
		FROM relation_types
		WHERE id = $1 AND deleted_at IS NULL
	`

	var rt knowledge.RelationType
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&rt.ID, &rt.WorkspaceID, &rt.Name, &rt.Slug, &rt.IsDirectional,
		&rt.CreatedAt, &rt.UpdatedAt, &rt.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("relation type not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get relation type: %w", err)
	}

	return &rt, nil
}

func (r *RelationTypeRepository) GetBySlug(ctx context.Context, workspaceID common.ID, slug common.Slug) (*knowledge.RelationType, error) {
	query := `
		SELECT id, workspace_id, name, slug, is_directional,
			created_at, updated_at, deleted_at
		FROM relation_types
		WHERE workspace_id = $1 AND slug = $2 AND deleted_at IS NULL
	`

	var rt knowledge.RelationType
	err := r.db.QueryRowContext(ctx, query, workspaceID, slug).Scan(
		&rt.ID, &rt.WorkspaceID, &rt.Name, &rt.Slug, &rt.IsDirectional,
		&rt.CreatedAt, &rt.UpdatedAt, &rt.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("relation type not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get relation type by slug: %w", err)
	}

	return &rt, nil
}

func (r *RelationTypeRepository) ListByWorkspace(ctx context.Context, workspaceID common.ID) ([]*knowledge.RelationType, error) {
	query := `
		SELECT id, workspace_id, name, slug, is_directional,
			created_at, updated_at, deleted_at
		FROM relation_types
		WHERE workspace_id = $1 AND deleted_at IS NULL
		ORDER BY name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to list relation types: %w", err)
	}
	defer rows.Close()

	var relationTypes []*knowledge.RelationType
	for rows.Next() {
		var rt knowledge.RelationType
		err := rows.Scan(
			&rt.ID, &rt.WorkspaceID, &rt.Name, &rt.Slug, &rt.IsDirectional,
			&rt.CreatedAt, &rt.UpdatedAt, &rt.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan relation type: %w", err)
		}
		relationTypes = append(relationTypes, &rt)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating relation types: %w", err)
	}

	return relationTypes, nil
}

func (r *RelationTypeRepository) Update(ctx context.Context, relationType *knowledge.RelationType) error {
	query := `
		UPDATE relation_types
		SET name = $2, slug = $3, is_directional = $4, updated_at = $5
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query,
		relationType.ID, relationType.Name, relationType.Slug,
		relationType.IsDirectional, relationType.UpdatedAt,
	)

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

func (r *RelationTypeRepository) Delete(ctx context.Context, id common.ID) error {
	query := `
		UPDATE relation_types
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, id)
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

// ===== Relation Repository =====

type RelationRepository struct {
	db *sql.DB
}

func NewRelationRepository(db *sql.DB) *RelationRepository {
	return &RelationRepository{db: db}
}

func (r *RelationRepository) Create(ctx context.Context, relation *knowledge.Relation) error {
	query := `
		INSERT INTO item_relations (
			id, workspace_id, from_item_id, to_item_id,
			relation_type_id, created_at, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.ExecContext(ctx, query,
		relation.ID, relation.WorkspaceID, relation.FromItemID, relation.ToItemID,
		relation.RelationTypeID, relation.CreatedAt, relation.CreatedBy,
	)

	if err != nil {
		return fmt.Errorf("failed to create relation: %w", err)
	}

	return nil
}

func (r *RelationRepository) GetByID(ctx context.Context, id common.ID) (*knowledge.Relation, error) {
	query := `
		SELECT r.id, r.workspace_id, r.from_item_id, r.to_item_id,
			r.relation_type_id, r.created_at, r.created_by, r.deleted_at,
			rt.id, rt.workspace_id, rt.name, rt.slug, rt.is_directional,
			rt.created_at, rt.updated_at
		FROM item_relations r
		JOIN relation_types rt ON r.relation_type_id = rt.id
		WHERE r.id = $1 AND r.deleted_at IS NULL
	`

	var rel knowledge.Relation
	var rt knowledge.RelationType
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&rel.ID, &rel.WorkspaceID, &rel.FromItemID, &rel.ToItemID,
		&rel.RelationTypeID, &rel.CreatedAt, &rel.CreatedBy, &rel.DeletedAt,
		&rt.ID, &rt.WorkspaceID, &rt.Name, &rt.Slug, &rt.IsDirectional,
		&rt.CreatedAt, &rt.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("relation not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get relation: %w", err)
	}

	rel.RelationType = &rt
	return &rel, nil
}

func (r *RelationRepository) GetByItems(ctx context.Context, fromItemID, toItemID, relationTypeID common.ID) (*knowledge.Relation, error) {
	query := `
		SELECT r.id, r.workspace_id, r.from_item_id, r.to_item_id,
			r.relation_type_id, r.created_at, r.created_by, r.deleted_at,
			rt.id, rt.workspace_id, rt.name, rt.slug, rt.is_directional,
			rt.created_at, rt.updated_at
		FROM item_relations r
		JOIN relation_types rt ON r.relation_type_id = rt.id
		WHERE r.from_item_id = $1 AND r.to_item_id = $2
			AND r.relation_type_id = $3 AND r.deleted_at IS NULL
	`

	var rel knowledge.Relation
	var rt knowledge.RelationType
	err := r.db.QueryRowContext(ctx, query, fromItemID, toItemID, relationTypeID).Scan(
		&rel.ID, &rel.WorkspaceID, &rel.FromItemID, &rel.ToItemID,
		&rel.RelationTypeID, &rel.CreatedAt, &rel.CreatedBy, &rel.DeletedAt,
		&rt.ID, &rt.WorkspaceID, &rt.Name, &rt.Slug, &rt.IsDirectional,
		&rt.CreatedAt, &rt.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("relation not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get relation by items: %w", err)
	}

	rel.RelationType = &rt
	return &rel, nil
}

func (r *RelationRepository) ListByFromItem(ctx context.Context, fromItemID common.ID) ([]*knowledge.Relation, error) {
	query := `
		SELECT r.id, r.workspace_id, r.from_item_id, r.to_item_id,
			r.relation_type_id, r.created_at, r.created_by, r.deleted_at,
			rt.id, rt.workspace_id, rt.name, rt.slug, rt.is_directional,
			rt.created_at, rt.updated_at
		FROM item_relations r
		JOIN relation_types rt ON r.relation_type_id = rt.id
		WHERE r.from_item_id = $1 AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC
	`

	return r.queryRelations(ctx, query, fromItemID)
}

func (r *RelationRepository) ListByToItem(ctx context.Context, toItemID common.ID) ([]*knowledge.Relation, error) {
	query := `
		SELECT r.id, r.workspace_id, r.from_item_id, r.to_item_id,
			r.relation_type_id, r.created_at, r.created_by, r.deleted_at,
			rt.id, rt.workspace_id, rt.name, rt.slug, rt.is_directional,
			rt.created_at, rt.updated_at
		FROM item_relations r
		JOIN relation_types rt ON r.relation_type_id = rt.id
		WHERE r.to_item_id = $1 AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC
	`

	return r.queryRelations(ctx, query, toItemID)
}

func (r *RelationRepository) ListByItem(ctx context.Context, itemID common.ID) ([]*knowledge.Relation, error) {
	query := `
		SELECT r.id, r.workspace_id, r.from_item_id, r.to_item_id,
			r.relation_type_id, r.created_at, r.created_by, r.deleted_at,
			rt.id, rt.workspace_id, rt.name, rt.slug, rt.is_directional,
			rt.created_at, rt.updated_at
		FROM item_relations r
		JOIN relation_types rt ON r.relation_type_id = rt.id
		WHERE (r.from_item_id = $1 OR r.to_item_id = $1) AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC
	`

	return r.queryRelations(ctx, query, itemID)
}

func (r *RelationRepository) ListByRelationType(ctx context.Context, relationTypeID common.ID) ([]*knowledge.Relation, error) {
	query := `
		SELECT r.id, r.workspace_id, r.from_item_id, r.to_item_id,
			r.relation_type_id, r.created_at, r.created_by, r.deleted_at,
			rt.id, rt.workspace_id, rt.name, rt.slug, rt.is_directional,
			rt.created_at, rt.updated_at
		FROM item_relations r
		JOIN relation_types rt ON r.relation_type_id = rt.id
		WHERE r.relation_type_id = $1 AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC
	`

	return r.queryRelations(ctx, query, relationTypeID)
}

func (r *RelationRepository) queryRelations(ctx context.Context, query string, arg common.ID) ([]*knowledge.Relation, error) {
	rows, err := r.db.QueryContext(ctx, query, arg)
	if err != nil {
		return nil, fmt.Errorf("failed to query relations: %w", err)
	}
	defer rows.Close()

	var relations []*knowledge.Relation
	for rows.Next() {
		var rel knowledge.Relation
		var rt knowledge.RelationType
		err := rows.Scan(
			&rel.ID, &rel.WorkspaceID, &rel.FromItemID, &rel.ToItemID,
			&rel.RelationTypeID, &rel.CreatedAt, &rel.CreatedBy, &rel.DeletedAt,
			&rt.ID, &rt.WorkspaceID, &rt.Name, &rt.Slug, &rt.IsDirectional,
			&rt.CreatedAt, &rt.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan relation: %w", err)
		}
		rel.RelationType = &rt
		relations = append(relations, &rel)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating relations: %w", err)
	}

	return relations, nil
}

func (r *RelationRepository) Delete(ctx context.Context, id common.ID) error {
	query := `
		UPDATE item_relations
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, id)
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

func (r *RelationRepository) HardDelete(ctx context.Context, id common.ID) error {
	query := `DELETE FROM item_relations WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to hard delete relation: %w", err)
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
