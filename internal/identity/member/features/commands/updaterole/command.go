package updaterole

import "errors"

var (
	ErrMemberNotFound = errors.New("member not found in this workspace")
	ErrInvalidRole    = errors.New("role must be reader, editor, or admin")
	ErrLastAdmin      = errors.New("cannot demote the last admin of the workspace")
)

type Command struct {
	WorkspaceID string
	UserID      string
	NewRole     string
}

type Result struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}
