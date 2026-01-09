package handlers

import (
	"net/http"

	"documind.jordi.org/internal/application/relation"
	"documind.jordi.org/internal/domain/common"
	"github.com/gin-gonic/gin"
)

type RelationHandler struct {
	service *relation.Service
}

func NewRelationHandler(service *relation.Service) *RelationHandler {
	return &RelationHandler{service: service}
}

// ===== Relation Type Requests/Responses =====

type CreateRelationTypeRequest struct {
	WorkspaceID   string `json:"workspace_id" binding:"required"`
	Name          string `json:"name" binding:"required"`
	Slug          string `json:"slug" binding:"required"`
	IsDirectional bool   `json:"is_directional"`
}

type UpdateRelationTypeRequest struct {
	Name          *string `json:"name"`
	Slug          *string `json:"slug"`
	IsDirectional *bool   `json:"is_directional"`
}

type RelationTypeResponse struct {
	ID            string `json:"id"`
	WorkspaceID   string `json:"workspace_id"`
	Name          string `json:"name"`
	Slug          string `json:"slug"`
	IsDirectional bool   `json:"is_directional"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

// ===== Relation Requests/Responses =====

type CreateRelationRequest struct {
	WorkspaceID    string `json:"workspace_id" binding:"required"`
	FromItemID     string `json:"from_item_id" binding:"required"`
	ToItemID       string `json:"to_item_id" binding:"required"`
	RelationTypeID string `json:"relation_type_id" binding:"required"`
	CreatedBy      string `json:"created_by" binding:"required"`
}

type RelationResponse struct {
	ID             string                `json:"id"`
	WorkspaceID    string                `json:"workspace_id"`
	FromItemID     string                `json:"from_item_id"`
	ToItemID       string                `json:"to_item_id"`
	RelationTypeID string                `json:"relation_type_id"`
	RelationType   *RelationTypeResponse `json:"relation_type,omitempty"`
	CreatedAt      string                `json:"created_at"`
	CreatedBy      string                `json:"created_by"`
}

// ===== Relation Type Handlers =====

func (h *RelationHandler) CreateRelationType(c *gin.Context) {
	var req CreateRelationTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := relation.CreateRelationTypeInput{
		WorkspaceID:   common.ID(req.WorkspaceID),
		Name:          req.Name,
		Slug:          common.Slug(req.Slug),
		IsDirectional: req.IsDirectional,
	}

	relationType, err := h.service.CreateRelationType(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, RelationTypeResponse{
		ID:            string(relationType.ID),
		WorkspaceID:   string(relationType.WorkspaceID),
		Name:          relationType.Name,
		Slug:          string(relationType.Slug),
		IsDirectional: relationType.IsDirectional,
		CreatedAt:     relationType.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     relationType.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

func (h *RelationHandler) GetRelationType(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	relationType, err := h.service.GetRelationType(c.Request.Context(), common.ID(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, RelationTypeResponse{
		ID:            string(relationType.ID),
		WorkspaceID:   string(relationType.WorkspaceID),
		Name:          relationType.Name,
		Slug:          string(relationType.Slug),
		IsDirectional: relationType.IsDirectional,
		CreatedAt:     relationType.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     relationType.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

func (h *RelationHandler) ListRelationTypes(c *gin.Context) {
	workspaceID := c.Query("workspace_id")
	if workspaceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspace_id is required"})
		return
	}

	relationTypes, err := h.service.ListRelationTypes(c.Request.Context(), common.ID(workspaceID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var response []RelationTypeResponse
	for _, rt := range relationTypes {
		response = append(response, RelationTypeResponse{
			ID:            string(rt.ID),
			WorkspaceID:   string(rt.WorkspaceID),
			Name:          rt.Name,
			Slug:          string(rt.Slug),
			IsDirectional: rt.IsDirectional,
			CreatedAt:     rt.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:     rt.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	c.JSON(http.StatusOK, response)
}

func (h *RelationHandler) UpdateRelationType(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	var req UpdateRelationTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := relation.UpdateRelationTypeInput{
		ID: common.ID(id),
	}

	if req.Name != nil {
		input.Name = req.Name
	}

	if req.Slug != nil {
		slug := common.Slug(*req.Slug)
		input.Slug = &slug
	}

	if req.IsDirectional != nil {
		input.IsDirectional = req.IsDirectional
	}

	relationType, err := h.service.UpdateRelationType(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, RelationTypeResponse{
		ID:            string(relationType.ID),
		WorkspaceID:   string(relationType.WorkspaceID),
		Name:          relationType.Name,
		Slug:          string(relationType.Slug),
		IsDirectional: relationType.IsDirectional,
		CreatedAt:     relationType.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     relationType.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

func (h *RelationHandler) DeleteRelationType(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	if err := h.service.DeleteRelationType(c.Request.Context(), common.ID(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// ===== Relation Handlers =====

func (h *RelationHandler) CreateRelation(c *gin.Context) {
	var req CreateRelationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := relation.CreateRelationInput{
		WorkspaceID:    common.ID(req.WorkspaceID),
		FromItemID:     common.ID(req.FromItemID),
		ToItemID:       common.ID(req.ToItemID),
		RelationTypeID: common.ID(req.RelationTypeID),
		CreatedBy:      common.ID(req.CreatedBy),
	}

	rel, err := h.service.CreateRelation(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := RelationResponse{
		ID:             string(rel.ID),
		WorkspaceID:    string(rel.WorkspaceID),
		FromItemID:     string(rel.FromItemID),
		ToItemID:       string(rel.ToItemID),
		RelationTypeID: string(rel.RelationTypeID),
		CreatedAt:      rel.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		CreatedBy:      string(rel.CreatedBy),
	}

	if rel.RelationType != nil {
		response.RelationType = &RelationTypeResponse{
			ID:            string(rel.RelationType.ID),
			WorkspaceID:   string(rel.RelationType.WorkspaceID),
			Name:          rel.RelationType.Name,
			Slug:          string(rel.RelationType.Slug),
			IsDirectional: rel.RelationType.IsDirectional,
			CreatedAt:     rel.RelationType.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:     rel.RelationType.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	c.JSON(http.StatusCreated, response)
}

func (h *RelationHandler) GetRelation(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	rel, err := h.service.GetRelation(c.Request.Context(), common.ID(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	response := RelationResponse{
		ID:             string(rel.ID),
		WorkspaceID:    string(rel.WorkspaceID),
		FromItemID:     string(rel.FromItemID),
		ToItemID:       string(rel.ToItemID),
		RelationTypeID: string(rel.RelationTypeID),
		CreatedAt:      rel.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		CreatedBy:      string(rel.CreatedBy),
	}

	if rel.RelationType != nil {
		response.RelationType = &RelationTypeResponse{
			ID:            string(rel.RelationType.ID),
			WorkspaceID:   string(rel.RelationType.WorkspaceID),
			Name:          rel.RelationType.Name,
			Slug:          string(rel.RelationType.Slug),
			IsDirectional: rel.RelationType.IsDirectional,
			CreatedAt:     rel.RelationType.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:     rel.RelationType.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	c.JSON(http.StatusOK, response)
}

func (h *RelationHandler) ListRelationsByItem(c *gin.Context) {
	itemID := c.Query("item_id")
	if itemID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "item_id is required"})
		return
	}

	relations, err := h.service.ListRelationsByItem(c.Request.Context(), common.ID(itemID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var response []RelationResponse
	for _, rel := range relations {
		r := RelationResponse{
			ID:             string(rel.ID),
			WorkspaceID:    string(rel.WorkspaceID),
			FromItemID:     string(rel.FromItemID),
			ToItemID:       string(rel.ToItemID),
			RelationTypeID: string(rel.RelationTypeID),
			CreatedAt:      rel.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			CreatedBy:      string(rel.CreatedBy),
		}

		if rel.RelationType != nil {
			r.RelationType = &RelationTypeResponse{
				ID:            string(rel.RelationType.ID),
				WorkspaceID:   string(rel.RelationType.WorkspaceID),
				Name:          rel.RelationType.Name,
				Slug:          string(rel.RelationType.Slug),
				IsDirectional: rel.RelationType.IsDirectional,
				CreatedAt:     rel.RelationType.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
				UpdatedAt:     rel.RelationType.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
			}
		}

		response = append(response, r)
	}

	c.JSON(http.StatusOK, response)
}

func (h *RelationHandler) DeleteRelation(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	if err := h.service.DeleteRelation(c.Request.Context(), common.ID(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
