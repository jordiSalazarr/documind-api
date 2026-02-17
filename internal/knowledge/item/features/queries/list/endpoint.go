package list

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	shareddomain "documind.jordi.org/internal/shared/domain"
)

// --- HTTP ---

const timeFormat = "2006-01-02T15:04:05Z07:00"

type itemListResponse struct {
	ID            string   `json:"id"`
	WorkspaceID   string   `json:"workspace_id"`
	ProjectID     string   `json:"project_id"`
	ServiceID     *string  `json:"service_id"`
	ItemTypeID    string   `json:"item_type_id"`
	Title         string   `json:"title"`
	Summary       string   `json:"summary"`
	Tags          []string `json:"tags"`
	LatestVersion int      `json:"latest_version"`
	Status        string   `json:"status"`
	OwnerUserID   string   `json:"owner_user_id"`
	CreatedAt     string   `json:"created_at"`
	UpdatedAt     string   `json:"updated_at"`
	CreatedBy     string   `json:"created_by"`
	UpdatedBy     string   `json:"updated_by"`
}

// Endpoint returns a gin.HandlerFunc for the list items endpoint.
func Endpoint(h *Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		projectID := c.Query("project_id")
		serviceID := c.Query("service_id")

		if projectID == "" && serviceID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "either project_id or service_id is required"})
			return
		}

		limit := parseIntParam(c.Query("limit"), defaultLimit)
		offset := parseIntParam(c.Query("offset"), 0)

		var result *ListResult
		var err error

		if serviceID != "" {
			result, err = h.HandleByService(c.Request.Context(), ByServiceQuery{
				ServiceID: shareddomain.ID(serviceID),
				Limit:     limit,
				Offset:    offset,
			})
		} else {
			result, err = h.HandleByProject(c.Request.Context(), ByProjectQuery{
				ProjectID: shareddomain.ID(projectID),
				Limit:     limit,
				Offset:    offset,
			})
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var response []itemListResponse
		for _, item := range result.Items {
			var svcID *string
			if item.ServiceID != nil {
				s := string(*item.ServiceID)
				svcID = &s
			}

			// Fetch the latest version to get title, summary, and tags
			var title, summary string
			var tags []string
			version, verErr := h.repo.GetLatestVersion(c.Request.Context(), item.ID)
			if verErr == nil && version != nil {
				title = version.Title
				summary = version.Summary
				tags = version.Tags
			}

			response = append(response, itemListResponse{
				ID:            string(item.ID),
				WorkspaceID:   string(item.WorkspaceID),
				ProjectID:     string(item.ProjectID),
				ServiceID:     svcID,
				ItemTypeID:    string(item.ItemTypeID),
				Title:         title,
				Summary:       summary,
				Tags:          tags,
				LatestVersion: item.LatestVersion,
				Status:        string(item.Status),
				OwnerUserID:   string(item.OwnerUserID),
				CreatedAt:     item.CreatedAt.Format(timeFormat),
				UpdatedAt:     item.UpdatedAt.Format(timeFormat),
				CreatedBy:     string(item.CreatedBy),
				UpdatedBy:     string(item.UpdatedBy),
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"items":       response,
			"total_count": result.TotalCount,
			"limit":       limit,
			"offset":      offset,
		})
	}
}

func parseIntParam(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return val
}
