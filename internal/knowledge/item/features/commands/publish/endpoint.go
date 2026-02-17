package publish

import (
	"net/http"

	"github.com/gin-gonic/gin"

	itemdomain "documind.jordi.org/internal/knowledge/domain"
	shareddomain "documind.jordi.org/internal/shared/domain"
)

// --- HTTP ---

const timeFormat = "2006-01-02T15:04:05Z07:00"

type itemResponse struct {
	ID            string  `json:"id"`
	WorkspaceID   string  `json:"workspace_id"`
	ProjectID     string  `json:"project_id"`
	ServiceID     *string `json:"service_id"`
	ItemTypeID    string  `json:"item_type_id"`
	LatestVersion int     `json:"latest_version"`
	Status        string  `json:"status"`
	OwnerUserID   string  `json:"owner_user_id"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
	CreatedBy     string  `json:"created_by"`
	UpdatedBy     string  `json:"updated_by"`
}

func toItemResponse(item *itemdomain.Item) itemResponse {
	var svcID *string
	if item.ServiceID != nil {
		s := string(*item.ServiceID)
		svcID = &s
	}

	return itemResponse{
		ID:            string(item.ID),
		WorkspaceID:   string(item.WorkspaceID),
		ProjectID:     string(item.ProjectID),
		ServiceID:     svcID,
		ItemTypeID:    string(item.ItemTypeID),
		LatestVersion: item.LatestVersion,
		Status:        string(item.Status),
		OwnerUserID:   string(item.OwnerUserID),
		CreatedAt:     item.CreatedAt.Format(timeFormat),
		UpdatedAt:     item.UpdatedAt.Format(timeFormat),
		CreatedBy:     string(item.CreatedBy),
		UpdatedBy:     string(item.UpdatedBy),
	}
}

// Endpoint returns a gin.HandlerFunc for the publish item endpoint.
func Endpoint(h *Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		updatedBy := c.Query("updated_by")

		if id == "" || updatedBy == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "id and updated_by are required"})
			return
		}

		updatedItem, err := h.Handle(c.Request.Context(), Command{
			ID:        shareddomain.ID(id),
			UpdatedBy: shareddomain.ID(updatedBy),
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, toItemResponse(updatedItem))
	}
}
