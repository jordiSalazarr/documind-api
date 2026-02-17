package domain

import (
	shared "documind.jordi.org/internal/shared/domain"
)

type Area struct {
	ID          shared.ID
	ProjectID   shared.ID
	Name        string
	Slug        shared.Slug
	Description string
	shared.Timestamp
}

func NewArea(projectID shared.ID, name string, slug shared.Slug, description string) *Area {
	return &Area{
		ID:          shared.NewID(),
		ProjectID:   projectID,
		Name:        name,
		Slug:        slug,
		Description: description,
		Timestamp:   shared.NewTimestamp(),
	}
}

func (a *Area) UpdateName(name string) {
	a.Name = name
	a.Update()
}

func (a *Area) UpdateDescription(description string) {
	a.Description = description
	a.Update()
}
