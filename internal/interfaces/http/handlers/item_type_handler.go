package handlers

import (
	"net/http"

	"documind.jordi.org/internal/application/itemtype"
	"documind.jordi.org/internal/domain/common"
	"github.com/gin-gonic/gin"
)

type ItemTypeHandler struct {
	service *itemtype.Service
}

func NewItemTypeHandler(service *itemtype.Service) *ItemTypeHandler {
	return &ItemTypeHandler{service: service}
}

type CreateItemTypeRequest struct {
	WorkspaceID string                 `json:"workspace_id" binding:"required"`
	Name        string                 `json:"name" binding:"required"`
	Slug        string                 `json:"slug" binding:"required"`
	SchemaJSON  map[string]interface{} `json:"schema_json"`
	Icon        *string                `json:"icon"`
	Color       *string                `json:"color"`
}

type UpdateItemTypeRequest struct {
	Name       string                 `json:"name" binding:"required"`
	SchemaJSON map[string]interface{} `json:"schema_json"`
	Icon       *string                `json:"icon"`
	Color      *string                `json:"color"`
}

type ItemTypeResponse struct {
	ID          string                 `json:"id"`
	WorkspaceID string                 `json:"workspace_id"`
	Name        string                 `json:"name"`
	Slug        string                 `json:"slug"`
	SchemaJSON  map[string]interface{} `json:"schema_json"`
	Icon        *string                `json:"icon"`
	Color       *string                `json:"color"`
	CreatedAt   string                 `json:"created_at"`
	UpdatedAt   string                 `json:"updated_at"`
}

func (h *ItemTypeHandler) Create(c *gin.Context) {
	var req CreateItemTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := itemtype.CreateItemTypeInput{
		WorkspaceID: common.ID(req.WorkspaceID),
		Name:        req.Name,
		Slug:        common.Slug(req.Slug),
		SchemaJSON:  req.SchemaJSON,
		Icon:        req.Icon,
		Color:       req.Color,
	}

	itemType, err := h.service.CreateItemType(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, ItemTypeResponse{
		ID:          string(itemType.ID),
		WorkspaceID: string(itemType.WorkspaceID),
		Name:        itemType.Name,
		Slug:        string(itemType.Slug),
		SchemaJSON:  itemType.SchemaJSON,
		Icon:        itemType.Icon,
		Color:       itemType.Color,
		CreatedAt:   itemType.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   itemType.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

func (h *ItemTypeHandler) Get(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	itemType, err := h.service.GetItemType(c.Request.Context(), common.ID(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ItemTypeResponse{
		ID:          string(itemType.ID),
		WorkspaceID: string(itemType.WorkspaceID),
		Name:        itemType.Name,
		Slug:        string(itemType.Slug),
		SchemaJSON:  itemType.SchemaJSON,
		Icon:        itemType.Icon,
		Color:       itemType.Color,
		CreatedAt:   itemType.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   itemType.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

func (h *ItemTypeHandler) GetBySlug(c *gin.Context) {
	workspaceID := c.Query("workspace_id")
	slug := c.Query("slug")

	if workspaceID == "" || slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspace_id and slug are required"})
		return
	}

	itemType, err := h.service.GetItemTypeBySlug(c.Request.Context(), common.ID(workspaceID), common.Slug(slug))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ItemTypeResponse{
		ID:          string(itemType.ID),
		WorkspaceID: string(itemType.WorkspaceID),
		Name:        itemType.Name,
		Slug:        string(itemType.Slug),
		SchemaJSON:  itemType.SchemaJSON,
		Icon:        itemType.Icon,
		Color:       itemType.Color,
		CreatedAt:   itemType.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   itemType.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

func (h *ItemTypeHandler) List(c *gin.Context) {
	workspaceID := c.Query("workspace_id")
	if workspaceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspace_id is required"})
		return
	}

	itemTypes, err := h.service.ListItemTypes(c.Request.Context(), common.ID(workspaceID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var response []ItemTypeResponse
	for _, itemType := range itemTypes {
		response = append(response, ItemTypeResponse{
			ID:          string(itemType.ID),
			WorkspaceID: string(itemType.WorkspaceID),
			Name:        itemType.Name,
			Slug:        string(itemType.Slug),
			SchemaJSON:  itemType.SchemaJSON,
			Icon:        itemType.Icon,
			Color:       itemType.Color,
			CreatedAt:   itemType.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:   itemType.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	c.JSON(http.StatusOK, response)
}

func (h *ItemTypeHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	var req UpdateItemTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := itemtype.UpdateItemTypeInput{
		ID:         common.ID(id),
		Name:       req.Name,
		SchemaJSON: req.SchemaJSON,
		Icon:       req.Icon,
		Color:      req.Color,
	}

	itemType, err := h.service.UpdateItemType(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ItemTypeResponse{
		ID:          string(itemType.ID),
		WorkspaceID: string(itemType.WorkspaceID),
		Name:        itemType.Name,
		Slug:        string(itemType.Slug),
		SchemaJSON:  itemType.SchemaJSON,
		Icon:        itemType.Icon,
		Color:       itemType.Color,
		CreatedAt:   itemType.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   itemType.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

func (h *ItemTypeHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	if err := h.service.DeleteItemType(c.Request.Context(), common.ID(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
