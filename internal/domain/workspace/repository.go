package workspace

import (
	"context"

	"documind.jordi.org/internal/domain/common"
)

// Repository defines the interface for workspace persistence
type Repository interface {
	Create(ctx context.Context, workspace *Workspace) error
	GetByID(ctx context.Context, id common.ID) (*Workspace, error)
	GetBySlug(ctx context.Context, slug common.Slug) (*Workspace, error)
	Update(ctx context.Context, workspace *Workspace) error
	Delete(ctx context.Context, id common.ID) error
	List(ctx context.Context, limit, offset int) ([]*Workspace, error)
}

// MemberRepository defines the interface for workspace member persistence
type MemberRepository interface {
	Create(ctx context.Context, member *Member) error
	GetByID(ctx context.Context, id common.ID) (*Member, error)
	GetByWorkspaceAndUser(ctx context.Context, workspaceID, userID common.ID) (*Member, error)
	ListByWorkspace(ctx context.Context, workspaceID common.ID) ([]*Member, error)
	Update(ctx context.Context, member *Member) error
	Delete(ctx context.Context, id common.ID) error
}
