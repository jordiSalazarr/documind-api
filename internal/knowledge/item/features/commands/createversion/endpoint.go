package createversion

import (
	"net/http"

	"github.com/gin-gonic/gin"

	itemdomain "documind.jordi.org/internal/knowledge/domain"
	shareddomain "documind.jordi.org/internal/shared/domain"
	"documind.jordi.org/internal/shared/infrastructure/middleware"
)

// --- HTTP ---

type createItemVersionRequest struct {
	Title        string                 `json:"title" binding:"required"`
	Summary      string                 `json:"summary"`
	BodyMd       string                 `json:"body_md"`
	CustomFields map[string]interface{} `json:"custom_fields"`
	Tags         []string               `json:"tags"`
}

const timeFormat = "2006-01-02T15:04:05Z07:00"

type itemVersionResponse struct {
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

func toItemVersionResponse(v *itemdomain.ItemVersion) itemVersionResponse {
	return itemVersionResponse{
		ID:           string(v.ID),
		ItemID:       string(v.ItemID),
		WorkspaceID:  string(v.WorkspaceID),
		Version:      v.Version,
		Title:        v.Title,
		Summary:      v.Summary,
		BodyMd:       v.BodyMd,
		CustomFields: v.CustomFields,
		Tags:         v.Tags,
		Status:       string(v.Status),
		CreatedAt:    v.CreatedAt.Format(timeFormat),
		CreatedBy:    string(v.CreatedBy),
	}
}

// Endpoint returns a gin.HandlerFunc for the create item version endpoint.
func Endpoint(h *Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		itemID := c.Param("id")
		if itemID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "item_id is required"})
			return
		}

		var req createItemVersionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Get workspace ID from the item
		item, err := h.repo.GetByID(c.Request.Context(), shareddomain.ID(itemID))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
			return
		}

		cmd := Command{
			ItemID:       shareddomain.ID(itemID),
			WorkspaceID:  item.WorkspaceID,
			Title:        req.Title,
			Summary:      req.Summary,
			BodyMd:       req.BodyMd,
			CustomFields: req.CustomFields,
			Tags:         req.Tags,
			CreatedBy:    shareddomain.ID(middleware.GetUserID(c)),
		}

		version, err := h.Handle(c.Request.Context(), cmd)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, toItemVersionResponse(version))
	}
}
