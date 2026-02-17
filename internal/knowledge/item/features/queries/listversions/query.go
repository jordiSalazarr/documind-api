package listversions

import (
	shareddomain "documind.jordi.org/internal/shared/domain"
)

// Query holds the parameters for listing all versions of an item.
type Query struct {
	ItemID shareddomain.ID
	Limit  int
	Offset int
}
