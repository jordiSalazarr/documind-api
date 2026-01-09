package knowledge

import (
	"documind.jordi.org/internal/domain/common"
)

// Project represents a logical grouping of items
type Project struct {
	ID          common.ID
	WorkspaceID common.ID
	Name        string
	Slug        common.Slug
	Description string
	common.Timestamp
	common.Audit
}

func NewProject(
	workspaceID common.ID,
	name string,
	slug common.Slug,
	description string,
	createdBy common.ID,
) *Project {
	return &Project{
		ID:          common.NewID(),
		WorkspaceID: workspaceID,
		Name:        name,
		Slug:        slug,
		Description: description,
		Timestamp:   common.NewTimestamp(),
		Audit:       common.NewAudit(createdBy),
	}
}

func (p *Project) Update(name, description string, updatedBy common.ID) {
	p.Name = name
	p.Description = description
	p.Audit.Update(updatedBy)
	p.Timestamp.Update()
}

// Service represents a service within a project
type Service struct {
	ID          common.ID
	WorkspaceID common.ID
	ProjectID   common.ID
	Name        string
	Slug        common.Slug
	Description string
	common.Timestamp
	common.Audit
}

func NewService(
	workspaceID, projectID common.ID,
	name string,
	slug common.Slug,
	description string,
	createdBy common.ID,
) *Service {
	return &Service{
		ID:          common.NewID(),
		WorkspaceID: workspaceID,
		ProjectID:   projectID,
		Name:        name,
		Slug:        slug,
		Description: description,
		Timestamp:   common.NewTimestamp(),
		Audit:       common.NewAudit(createdBy),
	}
}

func (s *Service) Update(name, description string, updatedBy common.ID) {
	s.Name = name
	s.Description = description
	s.Audit.Update(updatedBy)
	s.Timestamp.Update()
}
