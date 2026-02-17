package invite

import "errors"

var (
	ErrInvalidEmail  = errors.New("invalid email address")
	ErrInvalidRole   = errors.New("role must be reader, editor, or admin")
	ErrAlreadyMember = errors.New("user is already a member of this workspace")
	ErrUserNotFound  = errors.New("no user registered with this email")
)

type Command struct {
	WorkspaceID string
	Email       string
	Role        string
	InvitedBy   string
}

type Result struct {
	ID     string `json:"id"`
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
}
