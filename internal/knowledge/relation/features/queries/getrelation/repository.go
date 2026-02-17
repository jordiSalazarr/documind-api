package getrelation

import (
	"context"

	shared "documind.jordi.org/internal/shared/domain"
	reldomain "documind.jordi.org/internal/knowledge/domain"
)

type Repository interface {
	GetByID(ctx context.Context, id shared.ID) (*reldomain.Relation, error)
}
