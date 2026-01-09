package knowledge

import (
	"documind.jordi.org/internal/domain/common"
)

// ItemType defines configurable item types (Workflow, Decision, Rule, etc.)
type ItemType struct {
	ID          common.ID
	WorkspaceID common.ID
	Name        string
	Slug        common.Slug
	SchemaJSON  map[string]interface{} // JSON Schema for custom fields
	Icon        *string
	Color       *string
	common.Timestamp
}

func NewItemType(
	workspaceID common.ID,
	name string,
	slug common.Slug,
	schemaJSON map[string]interface{},
) *ItemType {
	return &ItemType{
		ID:          common.NewID(),
		WorkspaceID: workspaceID,
		Name:        name,
		Slug:        slug,
		SchemaJSON:  schemaJSON,
		Timestamp:   common.NewTimestamp(),
	}
}

// SetIcon sets the icon for the item type
func (it *ItemType) SetIcon(icon string) {
	it.Icon = &icon
}

// SetColor sets the color for the item type
func (it *ItemType) SetColor(color string) {
	it.Color = &color
}

// Update updates the item type fields
func (it *ItemType) Update(name string, schemaJSON map[string]interface{}) {
	it.Name = name
	it.SchemaJSON = schemaJSON
	it.Timestamp.Update()
}

// UpdateSchema updates the JSON schema
func (it *ItemType) UpdateSchema(schemaJSON map[string]interface{}) {
	it.SchemaJSON = schemaJSON
	it.Timestamp.Update()
}
