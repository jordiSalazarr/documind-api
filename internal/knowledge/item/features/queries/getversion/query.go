package getversion

import (
	shareddomain "documind.jordi.org/internal/shared/domain"
)

// Query holds the parameters for retrieving a specific item version.
type Query struct {
	ItemID  shareddomain.ID
	Version int
}
