package createrelation

import (
	"context"

	shared "documind.jordi.org/internal/shared/domain"
	reldomain "documind.jordi.org/internal/knowledge/domain"
)

type TypeRepository interface {
	GetByID(ctx context.Context, id shared.ID) (*reldomain.RelationType, error)
}

type RelationRepository interface {
	GetByID(ctx context.Context, id shared.ID) (*reldomain.Relation, error)
	Exists(ctx context.Context, fromItemID, toItemID, relationTypeID shared.ID) (bool, error)
	Create(ctx context.Context, relation *reldomain.Relation) error
}
