package publish

import (
	shareddomain "documind.jordi.org/internal/shared/domain"
)

// Command holds the data needed to publish an item.
type Command struct {
	ID        shareddomain.ID
	UpdatedBy shareddomain.ID
}
