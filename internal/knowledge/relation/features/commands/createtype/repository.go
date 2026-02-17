package createtype

import (
	"context"

	shared "documind.jordi.org/internal/shared/domain"
	reldomain "documind.jordi.org/internal/knowledge/domain"
)

type Repository interface {
	Create(ctx context.Context, rt *reldomain.RelationType) error
	Exists(ctx context.Context, workspaceID shared.ID, slug shared.Slug) (bool, error)
}
