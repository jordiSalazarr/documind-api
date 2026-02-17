package remove

import "errors"

var (
	ErrMemberNotFound = errors.New("member not found in this workspace")
	ErrLastAdmin      = errors.New("cannot remove the last admin of the workspace")
)

type Command struct {
	WorkspaceID string
	UserID      string
}
