package domain

import (
	shared "documind.jordi.org/internal/shared/domain"
)

type Project struct {
	ID          shared.ID
	WorkspaceID shared.ID
	Name        string
	Slug        shared.Slug
	Description string
	shared.Timestamp
	shared.Audit
}

func NewProject(workspaceID shared.ID, name string, slug shared.Slug, description string, createdBy shared.ID) *Project {
	return &Project{
		ID:          shared.NewID(),
		WorkspaceID: workspaceID,
		Name:        name,
		Slug:        slug,
		Description: description,
		Timestamp:   shared.NewTimestamp(),
		Audit:       shared.NewAudit(createdBy),
	}
}

func (p *Project) UpdateName(name string, updatedBy shared.ID) {
	p.Name = name
	p.Timestamp.Update()
	p.Audit.Update(updatedBy)
}

func (p *Project) UpdateDescription(description string, updatedBy shared.ID) {
	p.Description = description
	p.Timestamp.Update()
	p.Audit.Update(updatedBy)
}
