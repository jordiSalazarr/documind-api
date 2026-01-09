package knowledge

import (
	"context"
	"documind.jordi.org/internal/domain/common"
)

type ProjectRepository interface {
	Create(ctx context.Context, project *Project) error
	GetByID(ctx context.Context, id common.ID) (*Project, error)
	GetBySlug(ctx context.Context, workspaceID common.ID, slug common.Slug) (*Project, error)
	ListByWorkspace(ctx context.Context, workspaceID common.ID) ([]*Project, error)
	Update(ctx context.Context, project *Project) error
	Delete(ctx context.Context, id common.ID) error
	Exists(ctx context.Context, id common.ID) (bool, error)
}

type ServiceRepository interface {
	Create(ctx context.Context, service *Service) error
	GetByID(ctx context.Context, id common.ID) (*Service, error)
	GetBySlug(ctx context.Context, projectID common.ID, slug common.Slug) (*Service, error)
	ListByProject(ctx context.Context, projectID common.ID) ([]*Service, error)
	Update(ctx context.Context, service *Service) error
	Delete(ctx context.Context, id common.ID) error
	Exists(ctx context.Context, id common.ID) (bool, error)
}

type ItemTypeRepository interface {
	Create(ctx context.Context, itemType *ItemType) error
	GetByID(ctx context.Context, id common.ID) (*ItemType, error)
	GetBySlug(ctx context.Context, workspaceID common.ID, slug common.Slug) (*ItemType, error)
	ListByWorkspace(ctx context.Context, workspaceID common.ID) ([]*ItemType, error)
	Update(ctx context.Context, itemType *ItemType) error
	Delete(ctx context.Context, id common.ID) error
	Exists(ctx context.Context, id common.ID) (bool, error)
}

type ItemRepository interface {
	// Item operations
	CreateItem(ctx context.Context, item *Item) error
	GetItemByID(ctx context.Context, id common.ID) (*Item, error)
	ListItemsByProject(ctx context.Context, projectID common.ID) ([]*Item, error)
	ListItemsByService(ctx context.Context, serviceID common.ID) ([]*Item, error)
	UpdateItem(ctx context.Context, item *Item) error
	DeleteItem(ctx context.Context, id common.ID) error
	ItemExists(ctx context.Context, id common.ID) (bool, error)

	// ItemVersion operations
	CreateItemVersion(ctx context.Context, version *ItemVersion) error
	GetItemVersion(ctx context.Context, itemID common.ID, version int) (*ItemVersion, error)
	GetItemVersionByID(ctx context.Context, versionID common.ID) (*ItemVersion, error)
	GetLatestItemVersion(ctx context.Context, itemID common.ID) (*ItemVersion, error)
	ListItemVersions(ctx context.Context, itemID common.ID) ([]*ItemVersion, error)
	UpdateItemVersionStatus(ctx context.Context, itemID common.ID, version int, status ItemStatus) error
}
