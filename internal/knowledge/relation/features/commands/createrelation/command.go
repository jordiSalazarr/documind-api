package createrelation

import (
	"errors"
)

var (
	ErrRelationTypeNotFound  = errors.New("relation type not found")
	ErrRelationAlreadyExists = errors.New("relation already exists between these items")
	ErrSelfRelation          = errors.New("cannot create self-relation")
)

type Command struct {
	WorkspaceID    string
	FromItemID     string
	ToItemID       string
	RelationTypeID string
	CreatedBy      string
}

type RelationTypeResult struct {
	ID            string `json:"id"`
	WorkspaceID   string `json:"workspace_id"`
	Name          string `json:"name"`
	Slug          string `json:"slug"`
	IsDirectional bool   `json:"is_directional"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

type Result struct {
	ID             string              `json:"id"`
	WorkspaceID    string              `json:"workspace_id"`
	FromItemID     string              `json:"from_item_id"`
	ToItemID       string              `json:"to_item_id"`
	RelationTypeID string              `json:"relation_type_id"`
	RelationType   *RelationTypeResult `json:"relation_type,omitempty"`
	CreatedAt      string              `json:"created_at"`
	CreatedBy      string              `json:"created_by"`
}
