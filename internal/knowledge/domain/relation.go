package domain

import (
	"errors"
	"time"

	shared "documind.jordi.org/internal/shared/domain"
)

// RelationType defines configurable relation types
type RelationType struct {
	ID            shared.ID
	WorkspaceID   shared.ID
	Name          string
	Slug          shared.Slug
	IsDirectional bool
	shared.Timestamp
}

func NewRelationType(
	workspaceID shared.ID,
	name string,
	slug shared.Slug,
	isDirectional bool,
) *RelationType {
	return &RelationType{
		ID:            shared.NewID(),
		WorkspaceID:   workspaceID,
		Name:          name,
		Slug:          slug,
		IsDirectional: isDirectional,
		Timestamp:     shared.NewTimestamp(),
	}
}

// Relation represents a typed link between two items
type Relation struct {
	ID             shared.ID
	WorkspaceID    shared.ID
	FromItemID     shared.ID
	ToItemID       shared.ID
	RelationTypeID shared.ID
	RelationType   *RelationType
	CreatedAt      time.Time
	CreatedBy      shared.ID
	DeletedAt      *time.Time
}

func NewRelation(
	workspaceID, fromItemID, toItemID, relationTypeID, createdBy shared.ID,
) (*Relation, error) {
	if fromItemID == toItemID {
		return nil, errors.New("cannot create self-relation")
	}

	return &Relation{
		ID:             shared.NewID(),
		WorkspaceID:    workspaceID,
		FromItemID:     fromItemID,
		ToItemID:       toItemID,
		RelationTypeID: relationTypeID,
		CreatedAt:      time.Now(),
		CreatedBy:      createdBy,
	}, nil
}

// SoftDelete marks the relation as deleted
func (r *Relation) SoftDelete() {
	now := time.Now()
	r.DeletedAt = &now
}

// IsDeleted checks if the relation is deleted
func (r *Relation) IsDeleted() bool {
	return r.DeletedAt != nil
}
