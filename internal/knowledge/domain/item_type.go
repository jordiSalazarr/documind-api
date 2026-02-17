package domain

import (
	shareddomain "documind.jordi.org/internal/shared/domain"
)

// ItemType defines a configurable item type (Workflow, Decision, Rule, etc.).
type ItemType struct {
	ID          shareddomain.ID
	WorkspaceID shareddomain.ID
	Name        string
	Slug        shareddomain.Slug
	Description *string
	Icon        *string
	Fields      map[string]interface{}
	shareddomain.Timestamp
}

// NewItemType creates a new ItemType with generated ID and timestamps.
func NewItemType(
	workspaceID shareddomain.ID,
	name string,
	slug shareddomain.Slug,
	description *string,
	icon *string,
	fields map[string]interface{},
) *ItemType {
	return &ItemType{
		ID:          shareddomain.NewID(),
		WorkspaceID: workspaceID,
		Name:        name,
		Slug:        slug,
		Description: description,
		Icon:        icon,
		Fields:      fields,
		Timestamp:   shareddomain.NewTimestamp(),
	}
}

// Update modifies the mutable fields of an ItemType and refreshes UpdatedAt.
func (it *ItemType) Update(name string, description *string, icon *string, fields map[string]interface{}) {
	it.Name = name
	it.Description = description
	it.Icon = icon
	it.Fields = fields
	it.Timestamp.Update()
}
