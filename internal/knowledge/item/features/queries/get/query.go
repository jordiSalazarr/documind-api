package get

import (
	itemdomain "documind.jordi.org/internal/knowledge/domain"
	shareddomain "documind.jordi.org/internal/shared/domain"
)

// Query holds the parameters for retrieving an item.
type Query struct {
	ID shareddomain.ID
}

// Result holds the item together with its latest version.
type Result struct {
	Item    *itemdomain.Item
	Version *itemdomain.ItemVersion
}
