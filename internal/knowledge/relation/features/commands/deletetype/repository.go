package deletetype

import (
	"context"

	shared "documind.jordi.org/internal/shared/domain"
	reldomain "documind.jordi.org/internal/knowledge/domain"
)

type Repository interface {
	GetByID(ctx context.Context, id shared.ID) (*reldomain.RelationType, error)
	Delete(ctx context.Context, id shared.ID) error
}
