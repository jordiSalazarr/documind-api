package list

type Query struct {
	WorkspaceID string
}

type MemberResult struct {
	ID          string  `json:"id"`
	UserID      string  `json:"user_id"`
	Email       string  `json:"email"`
	Role        string  `json:"role"`
	InvitedAt   string  `json:"invited_at"`
	JoinedAt    *string `json:"joined_at"`
}
