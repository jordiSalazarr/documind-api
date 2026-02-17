package domain

import (
	"documind.jordi.org/internal/shared/domain"
)

// Workspace is the aggregate root for multi-tenant boundary
type Workspace struct {
	ID       domain.ID
	Name     string
	Slug     domain.Slug
	Settings map[string]interface{}
	domain.Timestamp
}

func NewWorkspace(name string, slug domain.Slug) *Workspace {
	return &Workspace{
		ID:        domain.NewID(),
		Name:      name,
		Slug:      slug,
		Settings:  make(map[string]interface{}),
		Timestamp: domain.NewTimestamp(),
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
	ID          domain.ID
	WorkspaceID domain.ID
	UserID      domain.ID
	Email       domain.Email
	Role        Role
	InvitedAt   domain.Timestamp
	JoinedAt    *domain.Timestamp
	InvitedBy   domain.ID
}

func NewMember(workspaceID, userID domain.ID, email domain.Email, role Role, invitedBy domain.ID) *Member {
	return &Member{
		ID:          domain.NewID(),
		WorkspaceID: workspaceID,
		UserID:      userID,
		Email:       email,
		Role:        role,
		InvitedAt:   domain.NewTimestamp(),
		InvitedBy:   invitedBy,
	}
}

func (m *Member) MarkAsJoined() {
	now := domain.NewTimestamp()
	m.JoinedAt = &now
}

func (m *Member) ChangeRole(newRole Role) {
	m.Role = newRole
}

func (m *Member) HasJoined() bool {
	return m.JoinedAt != nil
}
