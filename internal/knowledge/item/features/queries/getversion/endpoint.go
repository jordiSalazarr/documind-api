package getversion

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	itemdomain "documind.jordi.org/internal/knowledge/domain"
	shareddomain "documind.jordi.org/internal/shared/domain"
)

// --- HTTP ---

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

// Endpoint returns a gin.HandlerFunc for the get item version endpoint.
func Endpoint(h *Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		itemID := c.Param("id")
		versionStr := c.Param("versionId")

		if itemID == "" || versionStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "item_id and version are required"})
			return
		}

		versionNum, err := strconv.Atoi(versionStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid version number"})
			return
		}

		itemVersion, err := h.Handle(c.Request.Context(), Query{
			ItemID:  shareddomain.ID(itemID),
			Version: versionNum,
		})
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, toItemVersionResponse(itemVersion))
	}
}
