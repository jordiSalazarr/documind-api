package handlers

import (
	"net/http"

	"documind.jordi.org/internal/application/project"
	"documind.jordi.org/internal/domain/common"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AreaHandler struct {
	service *project.Service
}

func NewAreaHandler(service *project.Service) *AreaHandler {
	return &AreaHandler{service: service}
}

type CreateAreaRequest struct {
	WorkspaceID string `json:"workspace_id" binding:"required"`
	ProjectID   string `json:"project_id" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Slug        string `json:"slug" binding:"required"`
	Description string `json:"description"`
}

type UpdateAreaRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type AreaResponse struct {
	ID          string `json:"id"`
	WorkspaceID string `json:"workspace_id"`
	ProjectID   string `json:"project_id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	CreatedBy   string `json:"created_by"`
	UpdatedBy   string `json:"updated_by"`
}

func (h *AreaHandler) Create(c *gin.Context) {
	var req CreateAreaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	currentUserID := c.GetString("user_id")
	if currentUserID == "" {
		currentUserID = uuid.New().String()
	}

	input := project.CreateAreaInput{
		WorkspaceID: common.ID(req.WorkspaceID),
		ProjectID:   common.ID(req.ProjectID),
		Name:        req.Name,
		Slug:        common.Slug(req.Slug),
		Description: req.Description,
		CreatedBy:   common.ID(currentUserID),
	}

	service, err := h.service.CreateArea(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, AreaResponse{
		ID:          string(service.ID),
		WorkspaceID: string(service.WorkspaceID),
		ProjectID:   string(service.ProjectID),
		Name:        service.Name,
		Slug:        string(service.Slug),
		Description: service.Description,
		CreatedAt:   service.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   service.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		CreatedBy:   string(service.CreatedBy),
		UpdatedBy:   string(service.UpdatedBy),
	})
}

func (h *AreaHandler) Get(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	service, err := h.service.GetArea(c.Request.Context(), common.ID(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, AreaResponse{
		ID:          string(service.ID),
		WorkspaceID: string(service.WorkspaceID),
		ProjectID:   string(service.ProjectID),
		Name:        service.Name,
		Slug:        string(service.Slug),
		Description: service.Description,
		CreatedAt:   service.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   service.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		CreatedBy:   string(service.CreatedBy),
		UpdatedBy:   string(service.UpdatedBy),
	})
}

func (h *AreaHandler) GetBySlug(c *gin.Context) {
	projectID := c.Query("project_id")
	slug := c.Query("slug")

	if projectID == "" || slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project_id and slug are required"})
		return
	}

	service, err := h.service.GetAreaBySlug(c.Request.Context(), common.ID(projectID), common.Slug(slug))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, AreaResponse{
		ID:          string(service.ID),
		WorkspaceID: string(service.WorkspaceID),
		ProjectID:   string(service.ProjectID),
		Name:        service.Name,
		Slug:        string(service.Slug),
		Description: service.Description,
		CreatedAt:   service.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   service.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		CreatedBy:   string(service.CreatedBy),
		UpdatedBy:   string(service.UpdatedBy),
	})
}

func (h *AreaHandler) List(c *gin.Context) {
	projectID := c.Query("project_id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project_id is required"})
		return
	}

	areas, err := h.service.ListAreas(c.Request.Context(), common.ID(projectID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var response []AreaResponse
	for _, svc := range areas {
		response = append(response, AreaResponse{
			ID:          string(svc.ID),
			WorkspaceID: string(svc.WorkspaceID),
			ProjectID:   string(svc.ProjectID),
			Name:        svc.Name,
			Slug:        string(svc.Slug),
			Description: svc.Description,
			CreatedAt:   svc.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:   svc.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
			CreatedBy:   string(svc.CreatedBy),
			UpdatedBy:   string(svc.UpdatedBy),
		})
	}

	c.JSON(http.StatusOK, response)
}

func (h *AreaHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	var req UpdateAreaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	currentUserID := c.GetString("user_id")
	if currentUserID == "" {
		currentUserID = uuid.New().String()
	}

	input := project.UpdateAreaInput{
		ID:          common.ID(id),
		Name:        req.Name,
		Description: req.Description,
		UpdatedBy:   common.ID(currentUserID),
	}

	service, err := h.service.UpdateArea(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, AreaResponse{
		ID:          string(service.ID),
		WorkspaceID: string(service.WorkspaceID),
		ProjectID:   string(service.ProjectID),
		Name:        service.Name,
		Slug:        string(service.Slug),
		Description: service.Description,
		CreatedAt:   service.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   service.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		CreatedBy:   string(service.CreatedBy),
		UpdatedBy:   string(service.UpdatedBy),
	})
}

func (h *AreaHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	if err := h.service.DeleteArea(c.Request.Context(), common.ID(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
