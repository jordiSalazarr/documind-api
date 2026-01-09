package handlers

import (
	"net/http"
	"strconv"

	"documind.jordi.org/internal/application/search"
	"documind.jordi.org/internal/domain/common"
	"github.com/gin-gonic/gin"
)

type SearchHandler struct {
	service *search.Service
}

func NewSearchHandler(service *search.Service) *SearchHandler {
	return &SearchHandler{service: service}
}

type SearchResponse struct {
	Results    []ItemVersionResponse `json:"results"`
	TotalCount int                   `json:"total_count"`
	Limit      int                   `json:"limit"`
	Offset     int                   `json:"offset"`
	Query      string                `json:"query"`
}

func (h *SearchHandler) SearchItemVersions(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'q' is required"})
		return
	}

	workspaceID := c.Query("workspace_id")
	projectID := c.Query("project_id")

	if workspaceID == "" && projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "either workspace_id or project_id is required"})
		return
	}

	// Parse limit and offset
	limit := 20
	offset := 0

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Build search input
	input := search.SearchInput{
		Query:  query,
		Limit:  limit,
		Offset: offset,
	}

	if workspaceID != "" {
		wid := common.ID(workspaceID)
		input.WorkspaceID = &wid
	}

	if projectID != "" {
		pid := common.ID(projectID)
		input.ProjectID = &pid
	}

	// Execute search
	result, err := h.service.SearchItemVersions(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert to response
	var results []ItemVersionResponse
	for _, version := range result.Versions {
		results = append(results, ItemVersionResponse{
			ID:           string(version.ID),
			ItemID:       string(version.ItemID),
			WorkspaceID:  string(version.WorkspaceID),
			Version:      version.Version,
			Title:        version.Title,
			Summary:      version.Summary,
			BodyMd:       version.BodyMd,
			CustomFields: version.CustomFields,
			Tags:         version.Tags,
			Status:       string(version.Status),
			CreatedAt:    version.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			CreatedBy:    string(version.CreatedBy),
		})
	}

	c.JSON(http.StatusOK, SearchResponse{
		Results:    results,
		TotalCount: result.TotalCount,
		Limit:      result.Limit,
		Offset:     result.Offset,
		Query:      query,
	})
}

type SemanticSearchRequest struct {
	Embedding   []float32 `json:"embedding" binding:"required"`
	WorkspaceID *string   `json:"workspace_id"`
	ProjectID   *string   `json:"project_id"`
	Limit       int       `json:"limit"`
	Offset      int       `json:"offset"`
}

func (h *SearchHandler) SemanticSearch(c *gin.Context) {
	var req SemanticSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.WorkspaceID == nil && req.ProjectID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "either workspace_id or project_id is required"})
		return
	}

	// Build search input
	input := search.SemanticSearchInput{
		Embedding: req.Embedding,
		Limit:     req.Limit,
		Offset:    req.Offset,
	}

	if req.WorkspaceID != nil {
		wid := common.ID(*req.WorkspaceID)
		input.WorkspaceID = &wid
	}

	if req.ProjectID != nil {
		pid := common.ID(*req.ProjectID)
		input.ProjectID = &pid
	}

	// Execute search
	result, err := h.service.SearchItemVersionsBySemantic(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert to response
	var results []ItemVersionResponse
	for _, version := range result.Versions {
		results = append(results, ItemVersionResponse{
			ID:           string(version.ID),
			ItemID:       string(version.ItemID),
			WorkspaceID:  string(version.WorkspaceID),
			Version:      version.Version,
			Title:        version.Title,
			Summary:      version.Summary,
			BodyMd:       version.BodyMd,
			CustomFields: version.CustomFields,
			Tags:         version.Tags,
			Status:       string(version.Status),
			CreatedAt:    version.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			CreatedBy:    string(version.CreatedBy),
		})
	}

	c.JSON(http.StatusOK, SearchResponse{
		Results:    results,
		TotalCount: result.TotalCount,
		Limit:      result.Limit,
		Offset:     result.Offset,
		Query:      "semantic search",
	})
}

type UpdateEmbeddingRequest struct {
	VersionID string    `json:"version_id" binding:"required"`
	Embedding []float32 `json:"embedding" binding:"required"`
}

func (h *SearchHandler) UpdateEmbedding(c *gin.Context) {
	var req UpdateEmbeddingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UpdateEmbedding(c.Request.Context(), common.ID(req.VersionID), req.Embedding); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "embedding updated successfully"})
}
