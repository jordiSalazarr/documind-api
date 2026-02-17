package delete

import "errors"

var ErrWorkspaceNotFound = errors.New("workspace not found")

// --- Command ---

type Command struct {
	ID string
}
