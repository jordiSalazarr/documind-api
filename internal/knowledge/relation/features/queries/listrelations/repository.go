package listrelations

import (
	"context"

	shared "documind.jordi.org/internal/shared/domain"
	reldomain "documind.jordi.org/internal/knowledge/domain"
)

type Repository interface {
	ListByItem(ctx context.Context, itemID shared.ID, limit, offset int) ([]*reldomain.Relation, error)
}
