package delete

import (
	shareddomain "documind.jordi.org/internal/shared/domain"
)

// Command holds the data needed to delete an item.
type Command struct {
	ID shareddomain.ID
}
