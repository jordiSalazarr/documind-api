package updatetype

import (
	"context"

	shared "documind.jordi.org/internal/shared/domain"
	reldomain "documind.jordi.org/internal/knowledge/domain"
)

type Repository interface {
	GetByID(ctx context.Context, id shared.ID) (*reldomain.RelationType, error)
	Exists(ctx context.Context, workspaceID shared.ID, slug shared.Slug) (bool, error)
	Update(ctx context.Context, rt *reldomain.RelationType) error
}
