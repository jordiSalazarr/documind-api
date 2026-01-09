package workspace

import (
	"documind.jordi.org/internal/domain/common"
)

// Workspace is the aggregate root for multi-tenant boundary
type Workspace struct {
	ID       common.ID
	Name     string
	Slug     common.Slug
	Settings map[string]interface{}
	common.Timestamp
}

func NewWorkspace(name string, slug common.Slug) *Workspace {
	return &Workspace{
		ID:        common.NewID(),
		Name:      name,
		Slug:      slug,
		Settings:  make(map[string]interface{}),
		Timestamp: common.NewTimestamp(),
	}
}

func (w *Workspace) UpdateName(name string) {
	w.Name = name
	w.Update()
}

func (w *Workspace) UpdateSettings(settings map[string]interface{}) {
	w.Settings = settings
	w.Update()
}

// Role represents user roles in the workspace
type Role string

const (
	RoleReader Role = "reader"
	RoleEditor Role = "editor"
	RoleAdmin  Role = "admin"
)

func (r Role) String() string {
	return string(r)
}

func (r Role) CanWrite() bool {
	return r == RoleEditor || r == RoleAdmin
}

func (r Role) CanManageMembers() bool {
	return r == RoleAdmin
}

// Member represents a workspace member
type Member struct {
	ID          common.ID
	WorkspaceID common.ID
	UserID      common.ID
	Email       common.Email
	Role        Role
	InvitedAt   common.Timestamp
	JoinedAt    *common.Timestamp
	InvitedBy   common.ID
}

func NewMember(workspaceID, userID common.ID, email common.Email, role Role, invitedBy common.ID) *Member {
	return &Member{
		ID:          common.NewID(),
		WorkspaceID: workspaceID,
		UserID:      userID,
		Email:       email,
		Role:        role,
		InvitedAt:   common.NewTimestamp(),
		InvitedBy:   invitedBy,
	}
}

func (m *Member) MarkAsJoined() {
	now := common.NewTimestamp()
	m.JoinedAt = &now
}

func (m *Member) ChangeRole(newRole Role) {
	m.Role = newRole
}

func (m *Member) HasJoined() bool {
	return m.JoinedAt != nil
}
