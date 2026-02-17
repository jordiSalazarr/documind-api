package revoke

import "errors"

var (
	ErrKeyNotFound = errors.New("API key not found")
)

// Command contains parameters for revoking an API key.
type Command struct {
	ID          string
	WorkspaceID string
}
