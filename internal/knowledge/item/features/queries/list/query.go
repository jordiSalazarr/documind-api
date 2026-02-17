package list

import (
	itemdomain "documind.jordi.org/internal/knowledge/domain"
	shareddomain "documind.jordi.org/internal/shared/domain"
)

// ByProjectQuery holds the parameters for listing items by project.
type ByProjectQuery struct {
	ProjectID shareddomain.ID
	Limit     int
	Offset    int
}

// ByServiceQuery holds the parameters for listing items by service.
type ByServiceQuery struct {
	ServiceID shareddomain.ID
	Limit     int
	Offset    int
}

// ListResult holds the paginated list of items along with total count.
type ListResult struct {
	Items      []*itemdomain.Item
	TotalCount int
}
