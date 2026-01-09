package handlers

import (
	"net/http"

	"documind.jordi.org/internal/application/project"
	"documind.jordi.org/internal/domain/common"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ProjectHandler struct {
	service *project.Service
}

func NewProjectHandler(service *project.Service) *ProjectHandler {
	return &ProjectHandler{service: service}
}

type CreateProjectRequest struct {
	WorkspaceID string `json:"workspace_id" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Slug        string `json:"slug" binding:"required"`
	Description string `json:"description"`
}

type UpdateProjectRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type ProjectResponse struct {
	ID          string `json:"id"`
	WorkspaceID string `json:"workspace_id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	CreatedBy   string `json:"created_by"`
	UpdatedBy   string `json:"updated_by"`
}

func (h *ProjectHandler) Create(c *gin.Context) {
	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	currentUserID := c.GetString("user_id")
	if currentUserID == "" {
		currentUserID = uuid.New().String()
	}

	input := project.CreateProjectInput{
		WorkspaceID: common.ID(req.WorkspaceID),
		Name:        req.Name,
		Slug:        common.Slug(req.Slug),
		Description: req.Description,
		CreatedBy:   common.ID(currentUserID),
	}

	proj, err := h.service.CreateProject(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, ProjectResponse{
		ID:          string(proj.ID),
		WorkspaceID: string(proj.WorkspaceID),
		Name:        proj.Name,
		Slug:        string(proj.Slug),
		Description: proj.Description,
		CreatedAt:   proj.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   proj.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		CreatedBy:   string(proj.CreatedBy),
		UpdatedBy:   string(proj.UpdatedBy),
	})
}

func (h *ProjectHandler) Get(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	proj, err := h.service.GetProject(c.Request.Context(), common.ID(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ProjectResponse{
		ID:          string(proj.ID),
		WorkspaceID: string(proj.WorkspaceID),
		Name:        proj.Name,
		Slug:        string(proj.Slug),
		Description: proj.Description,
		CreatedAt:   proj.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   proj.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		CreatedBy:   string(proj.CreatedBy),
		UpdatedBy:   string(proj.UpdatedBy),
	})
}

func (h *ProjectHandler) GetBySlug(c *gin.Context) {
	workspaceID := c.Query("workspace_id")
	slug := c.Query("slug")

	if workspaceID == "" || slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspace_id and slug are required"})
		return
	}

	proj, err := h.service.GetProjectBySlug(c.Request.Context(), common.ID(workspaceID), common.Slug(slug))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ProjectResponse{
		ID:          string(proj.ID),
		WorkspaceID: string(proj.WorkspaceID),
		Name:        proj.Name,
		Slug:        string(proj.Slug),
		Description: proj.Description,
		CreatedAt:   proj.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   proj.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		CreatedBy:   string(proj.CreatedBy),
		UpdatedBy:   string(proj.UpdatedBy),
	})
}

func (h *ProjectHandler) List(c *gin.Context) {
	workspaceID := c.Query("workspace_id")
	if workspaceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspace_id is required"})
		return
	}

	projects, err := h.service.ListProjects(c.Request.Context(), common.ID(workspaceID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var response []ProjectResponse
	for _, proj := range projects {
		response = append(response, ProjectResponse{
			ID:          string(proj.ID),
			WorkspaceID: string(proj.WorkspaceID),
			Name:        proj.Name,
			Slug:        string(proj.Slug),
			Description: proj.Description,
			CreatedAt:   proj.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:   proj.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
			CreatedBy:   string(proj.CreatedBy),
			UpdatedBy:   string(proj.UpdatedBy),
		})
	}

	c.JSON(http.StatusOK, response)
}

func (h *ProjectHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	var req UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	currentUserID := c.GetString("user_id")
	if currentUserID == "" {
		currentUserID = uuid.New().String()
	}

	input := project.UpdateProjectInput{
		ID:          common.ID(id),
		Name:        req.Name,
		Description: req.Description,
		UpdatedBy:   common.ID(currentUserID),
	}

	proj, err := h.service.UpdateProject(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ProjectResponse{
		ID:          string(proj.ID),
		WorkspaceID: string(proj.WorkspaceID),
		Name:        proj.Name,
		Slug:        string(proj.Slug),
		Description: proj.Description,
		CreatedAt:   proj.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   proj.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		CreatedBy:   string(proj.CreatedBy),
		UpdatedBy:   string(proj.UpdatedBy),
	})
}

func (h *ProjectHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	if err := h.service.DeleteProject(c.Request.Context(), common.ID(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
