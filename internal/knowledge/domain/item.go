package domain

import (
	"time"

	shareddomain "documind.jordi.org/internal/shared/domain"
)

// ItemStatus represents the status of an item
type ItemStatus string

const (
	ItemStatusDraft      ItemStatus = "draft"
	ItemStatusPublished  ItemStatus = "published"
	ItemStatusDeprecated ItemStatus = "deprecated"
)

// Item represents a knowledge item (workflow, decision, rule, etc.)
type Item struct {
	ID            shareddomain.ID
	WorkspaceID   shareddomain.ID
	ProjectID     shareddomain.ID
	ServiceID     *shareddomain.ID // Optional service
	ItemTypeID    shareddomain.ID
	LatestVersion int
	Status        ItemStatus
	OwnerUserID   shareddomain.ID
	shareddomain.Timestamp
	shareddomain.Audit
}

// ItemVersion represents a specific version of an item's content
type ItemVersion struct {
	ID           shareddomain.ID
	ItemID       shareddomain.ID
	WorkspaceID  shareddomain.ID
	Version      int
	Title        string
	Summary      string
	BodyMd       string                 // Markdown content
	CustomFields map[string]interface{} // Custom fields based on ItemType schema
	Tags         []string
	Status       ItemStatus
	CreatedAt    time.Time // when this version was created
	CreatedBy    shareddomain.ID // who created this version
}

func NewItem(
	workspaceID shareddomain.ID,
	projectID shareddomain.ID,
	itemTypeID shareddomain.ID,
	ownerUserID shareddomain.ID,
	createdBy shareddomain.ID,
) *Item {
	return &Item{
		ID:            shareddomain.NewID(),
		WorkspaceID:   workspaceID,
		ProjectID:     projectID,
		ItemTypeID:    itemTypeID,
		LatestVersion: 0, // Will increment when first version is created
		Status:        ItemStatusDraft,
		OwnerUserID:   ownerUserID,
		Timestamp:     shareddomain.NewTimestamp(),
		Audit:         shareddomain.NewAudit(createdBy),
	}
}

func NewItemVersion(
	itemID shareddomain.ID,
	workspaceID shareddomain.ID,
	version int,
	title string,
	summary string,
	bodyMd string,
	customFields map[string]interface{},
	tags []string,
	createdBy shareddomain.ID,
) *ItemVersion {
	if customFields == nil {
		customFields = make(map[string]interface{})
	}
	if tags == nil {
		tags = []string{}
	}

	return &ItemVersion{
		ID:           shareddomain.NewID(),
		ItemID:       itemID,
		WorkspaceID:  workspaceID,
		Version:      version,
		Title:        title,
		Summary:      summary,
		BodyMd:       bodyMd,
		CustomFields: customFields,
		Tags:         tags,
		Status:       ItemStatusDraft,
		CreatedAt:    time.Now(),
		CreatedBy:    createdBy,
	}
}

// SetService assigns the item to a specific service within the project
func (i *Item) SetService(serviceID shareddomain.ID) {
	i.ServiceID = &serviceID
}

// ClearService removes the service assignment
func (i *Item) ClearService() {
	i.ServiceID = nil
}

// Publish changes the status to published
func (i *Item) Publish(updatedBy shareddomain.ID) {
	i.Status = ItemStatusPublished
	i.Audit.Update(updatedBy)
	i.Timestamp.Update()
}

// Deprecate changes the status to deprecated
func (i *Item) Deprecate(updatedBy shareddomain.ID) {
	i.Status = ItemStatusDeprecated
	i.Audit.Update(updatedBy)
	i.Timestamp.Update()
}

// IncrementVersion increments the latest version number
func (i *Item) IncrementVersion(updatedBy shareddomain.ID) {
	i.LatestVersion++
	i.Audit.Update(updatedBy)
	i.Timestamp.Update()
}

// Publish changes the version status to published
func (v *ItemVersion) Publish() {
	v.Status = ItemStatusPublished
}

// Deprecate changes the version status to deprecated
func (v *ItemVersion) Deprecate() {
	v.Status = ItemStatusDeprecated
}
