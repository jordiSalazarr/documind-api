package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	workspaceapp "documind.jordi.org/internal/application/workspace"
)

type WorkspaceHandler struct {
	service *workspaceapp.Service
}

func NewWorkspaceHandler(service *workspaceapp.Service) *WorkspaceHandler {
	return &WorkspaceHandler{
		service: service,
	}
}

type CreateWorkspaceRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug,omitempty"`
}

type UpdateWorkspaceRequest struct {
	Name     string                 `json:"name,omitempty"`
	Settings map[string]interface{} `json:"settings,omitempty"`
}

func (h *WorkspaceHandler) CreateWorkspace(c *gin.Context) {
	var req CreateWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "message": err.Error()})
		return
	}

	cmd := workspaceapp.CreateWorkspaceCommand{
		Name: req.Name,
		Slug: req.Slug,
	}

	workspace, err := h.service.CreateWorkspace(c.Request.Context(), cmd)
	if err != nil {
		switch {
		case errors.Is(err, workspaceapp.ErrInvalidWorkspaceName):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workspace name", "message": err.Error()})
		case errors.Is(err, workspaceapp.ErrInvalidSlug):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid slug", "message": err.Error()})
		case errors.Is(err, workspaceapp.ErrWorkspaceExists):
			c.JSON(http.StatusConflict, gin.H{"error": "workspace already exists", "message": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create workspace", "message": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, workspace)
}

func (h *WorkspaceHandler) GetWorkspace(c *gin.Context) {
	id := c.Param("id")

	workspace, err := h.service.GetWorkspace(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, workspaceapp.ErrWorkspaceNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "workspace not found", "message": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get workspace", "message": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, workspace)
}

func (h *WorkspaceHandler) GetWorkspaceBySlug(c *gin.Context) {
	slug := c.Param("slug")

	workspace, err := h.service.GetWorkspaceBySlug(c.Request.Context(), slug)
	if err != nil {
		if errors.Is(err, workspaceapp.ErrWorkspaceNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "workspace not found", "message": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get workspace", "message": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, workspace)
}

func (h *WorkspaceHandler) UpdateWorkspace(c *gin.Context) {
	id := c.Param("id")

	var req UpdateWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "message": err.Error()})
		return
	}

	cmd := workspaceapp.UpdateWorkspaceCommand{
		ID:       id,
		Name:     req.Name,
		Settings: req.Settings,
	}

	workspace, err := h.service.UpdateWorkspace(c.Request.Context(), cmd)
	if err != nil {
		if errors.Is(err, workspaceapp.ErrWorkspaceNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "workspace not found", "message": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update workspace", "message": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, workspace)
}

func (h *WorkspaceHandler) DeleteWorkspace(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.DeleteWorkspace(c.Request.Context(), id); err != nil {
		if errors.Is(err, workspaceapp.ErrWorkspaceNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "workspace not found", "message": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete workspace", "message": err.Error()})
		}
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *WorkspaceHandler) ListWorkspaces(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	workspaces, err := h.service.ListWorkspaces(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list workspaces", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"workspaces": workspaces,
		"limit":      limit,
		"offset":     offset,
	})
}
