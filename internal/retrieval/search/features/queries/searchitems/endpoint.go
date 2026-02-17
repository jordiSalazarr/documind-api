package searchitems

import (
	"net/http"
	"strconv"

	shared "documind.jordi.org/internal/shared/domain"
	"documind.jordi.org/internal/shared/infrastructure/middleware"
	"github.com/gin-gonic/gin"
)

// SearchResponse represents the JSON response for search endpoints
type SearchResponse struct {
	Results    []ItemVersionResponse `json:"results"`
	TotalCount int                   `json:"total_count"`
	Limit      int                   `json:"limit"`
	Offset     int                   `json:"offset"`
	Query      string                `json:"query"`
}

// ItemVersionResponse represents an item version in API responses
type ItemVersionResponse struct {
	ID           string                 `json:"id"`
	ItemID       string                 `json:"item_id"`
	WorkspaceID  string                 `json:"workspace_id"`
	Version      int                    `json:"version"`
	Title        string                 `json:"title"`
	Summary      string                 `json:"summary"`
	BodyMd       string                 `json:"body_md"`
	CustomFields map[string]interface{} `json:"custom_fields"`
	Tags         []string               `json:"tags"`
	Status       string                 `json:"status"`
	CreatedAt    string                 `json:"created_at"`
	CreatedBy    string                 `json:"created_by"`
}

// SemanticSearchRequest represents the JSON body for semantic search
type SemanticSearchRequest struct {
	Embedding []float32 `json:"embedding" binding:"required"`
	ProjectID *string   `json:"project_id"`
	Limit     int       `json:"limit"`
	Offset    int       `json:"offset"`
}

// Endpoint returns a gin.HandlerFunc for GET /search/items
func Endpoint(h *Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		q := c.DefaultQuery("q", "")
		if q == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'q' is required"})
			return
		}

		workspaceID := middleware.GetWorkspaceID(c)
		projectID := c.Query("project_id")

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

		// Build search query
		wid := shared.ID(workspaceID)
		query := Query{
			QueryText:   q,
			WorkspaceID: &wid,
			Limit:       limit,
			Offset:      offset,
		}

		if projectID != "" {
			pid := shared.ID(projectID)
			query.ProjectID = &pid
		}

		// Execute search
		result, err := h.Handle(c.Request.Context(), query)
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
				Status:       version.Status,
				CreatedAt:    version.CreatedAt,
				CreatedBy:    string(version.CreatedBy),
			})
		}

		c.JSON(http.StatusOK, SearchResponse{
			Results:    results,
			TotalCount: result.TotalCount,
			Limit:      result.Limit,
			Offset:     result.Offset,
			Query:      q,
		})
	}
}

// SemanticEndpoint returns a gin.HandlerFunc for POST /search/semantic
func SemanticEndpoint(h *Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req SemanticSearchRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		workspaceID := middleware.GetWorkspaceID(c)

		// Build semantic query
		wid := shared.ID(workspaceID)
		query := SemanticQuery{
			Embedding:   req.Embedding,
			WorkspaceID: &wid,
			Limit:       req.Limit,
			Offset:      req.Offset,
		}

		if req.ProjectID != nil {
			pid := shared.ID(*req.ProjectID)
			query.ProjectID = &pid
		}

		// Execute search
		result, err := h.HandleSemantic(c.Request.Context(), query)
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
				Status:       version.Status,
				CreatedAt:    version.CreatedAt,
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
}
