package listtypes

import (
	"context"

	shared "documind.jordi.org/internal/shared/domain"
	reldomain "documind.jordi.org/internal/knowledge/domain"
)

type Repository interface {
	ListByWorkspace(ctx context.Context, workspaceID shared.ID, limit, offset int) ([]*reldomain.RelationType, error)
}
