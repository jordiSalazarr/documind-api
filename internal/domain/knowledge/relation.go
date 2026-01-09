package knowledge

import (
	"documind.jordi.org/internal/domain/common"
	"errors"
	"time"
)

// RelationType defines configurable relation types
type RelationType struct {
	ID             common.ID
	WorkspaceID    common.ID
	Name           string
	Slug           common.Slug
	IsDirectional  bool
	common.Timestamp
}

func NewRelationType(
	workspaceID common.ID,
	name string,
	slug common.Slug,
	isDirectional bool,
) *RelationType {
	return &RelationType{
		ID:            common.NewID(),
		WorkspaceID:   workspaceID,
		Name:          name,
		Slug:          slug,
		IsDirectional: isDirectional,
		Timestamp:     common.NewTimestamp(),
	}
}

// Relation represents a typed link between two items
type Relation struct {
	ID             common.ID
	WorkspaceID    common.ID
	FromItemID     common.ID
	ToItemID       common.ID
	RelationTypeID common.ID
	RelationType   *RelationType
	CreatedAt      time.Time
	CreatedBy      common.ID
	DeletedAt      *time.Time
}

func NewRelation(
	workspaceID, fromItemID, toItemID, relationTypeID, createdBy common.ID,
) (*Relation, error) {
	if fromItemID == toItemID {
		return nil, errors.New("cannot create self-relation")
	}

	return &Relation{
		ID:             common.NewID(),
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
