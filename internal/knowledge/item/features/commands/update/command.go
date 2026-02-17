package update

import (
	shareddomain "documind.jordi.org/internal/shared/domain"
)

// Command holds the data needed to update item metadata.
type Command struct {
	ID          shareddomain.ID
	ProjectID   *shareddomain.ID
	ServiceID   *shareddomain.ID
	OwnerUserID *shareddomain.ID
	UpdatedBy   shareddomain.ID
}
