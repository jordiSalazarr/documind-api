package knowledge

import (
	"context"

	"documind.jordi.org/internal/domain/common"
)

// RelationTypeRepository defines operations for relation types
type RelationTypeRepository interface {
	Create(ctx context.Context, relationType *RelationType) error
	GetByID(ctx context.Context, id common.ID) (*RelationType, error)
	GetBySlug(ctx context.Context, workspaceID common.ID, slug common.Slug) (*RelationType, error)
	ListByWorkspace(ctx context.Context, workspaceID common.ID) ([]*RelationType, error)
	Update(ctx context.Context, relationType *RelationType) error
	Delete(ctx context.Context, id common.ID) error
}

// RelationRepository defines operations for relations between items
type RelationRepository interface {
	Create(ctx context.Context, relation *Relation) error
	GetByID(ctx context.Context, id common.ID) (*Relation, error)
	GetByItems(ctx context.Context, fromItemID, toItemID, relationTypeID common.ID) (*Relation, error)
	ListByFromItem(ctx context.Context, fromItemID common.ID) ([]*Relation, error)
	ListByToItem(ctx context.Context, toItemID common.ID) ([]*Relation, error)
	ListByItem(ctx context.Context, itemID common.ID) ([]*Relation, error)
	ListByRelationType(ctx context.Context, relationTypeID common.ID) ([]*Relation, error)
	Delete(ctx context.Context, id common.ID) error
	HardDelete(ctx context.Context, id common.ID) error
}
