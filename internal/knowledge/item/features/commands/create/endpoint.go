package create

import (
	"net/http"

	"github.com/gin-gonic/gin"

	itemdomain "documind.jordi.org/internal/knowledge/domain"
	shareddomain "documind.jordi.org/internal/shared/domain"
	"documind.jordi.org/internal/shared/infrastructure/middleware"
)

// --- HTTP ---

type createItemRequest struct {
	ProjectID    string                 `json:"project_id" binding:"required"`
	ServiceID    *string                `json:"service_id"`
	ItemTypeID   string                 `json:"item_type_id" binding:"required"`
	Title        string                 `json:"title" binding:"required"`
	Summary      string                 `json:"summary"`
	BodyMd       string                 `json:"body_md"`
	CustomFields map[string]interface{} `json:"custom_fields"`
	Tags         []string               `json:"tags"`
}

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

type itemWithVersionResponse struct {
	Item    itemResponse        `json:"item"`
	Version itemVersionResponse `json:"version"`
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

// Endpoint returns a gin.HandlerFunc for the create item endpoint.
func Endpoint(h *Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req createItemRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userID := shareddomain.ID(middleware.GetUserID(c))
		workspaceID := shareddomain.ID(middleware.GetWorkspaceID(c))

		var serviceID *shareddomain.ID
		if req.ServiceID != nil {
			id := shareddomain.ID(*req.ServiceID)
			serviceID = &id
		}

		cmd := Command{
			WorkspaceID:  workspaceID,
			ProjectID:    shareddomain.ID(req.ProjectID),
			ServiceID:    serviceID,
			ItemTypeID:   shareddomain.ID(req.ItemTypeID),
			Title:        req.Title,
			Summary:      req.Summary,
			BodyMd:       req.BodyMd,
			CustomFields: req.CustomFields,
			Tags:         req.Tags,
			OwnerUserID:  userID,
			CreatedBy:    userID,
		}

		result, err := h.Handle(c.Request.Context(), cmd)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, itemWithVersionResponse{
			Item:    toItemResponse(result.Item),
			Version: toItemVersionResponse(result.Version),
		})
	}
}
