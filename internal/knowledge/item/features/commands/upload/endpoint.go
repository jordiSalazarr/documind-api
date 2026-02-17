package upload

import (
	nethttp "net/http"

	shareddomain "documind.jordi.org/internal/shared/domain"
	"documind.jordi.org/internal/shared/infrastructure/middleware"
	"github.com/gin-gonic/gin"
)

// uploadResponse is the HTTP response for a successful upload.
type uploadResponse struct {
	Item    itemResponse    `json:"item"`
	Version versionResponse `json:"version"`
	Parsing ParsingResult   `json:"parsing"`
}

type itemResponse struct {
	ID            string `json:"id"`
	WorkspaceID   string `json:"workspace_id"`
	ProjectID     string `json:"project_id"`
	ItemTypeID    string `json:"item_type_id"`
	LatestVersion int    `json:"latest_version"`
	Status        string `json:"status"`
	CreatedAt     string `json:"created_at"`
}

type versionResponse struct {
	ID      string `json:"id"`
	ItemID  string `json:"item_id"`
	Version int    `json:"version"`
	Title   string `json:"title"`
}

// Endpoint returns a gin handler for file upload.
func Endpoint(h *Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID := middleware.GetWorkspaceID(c)
		userID := middleware.GetUserID(c)

		// Required form fields
		projectID := c.PostForm("project_id")
		if projectID == "" {
			c.JSON(nethttp.StatusBadRequest, gin.H{"error": "project_id is required"})
			return
		}

		itemTypeID := c.PostForm("item_type_id")
		if itemTypeID == "" {
			c.JSON(nethttp.StatusBadRequest, gin.H{"error": "item_type_id is required"})
			return
		}

		// File
		fileHeader, err := c.FormFile("file")
		if err != nil {
			c.JSON(nethttp.StatusBadRequest, gin.H{"error": "file is required"})
			return
		}

		file, err := fileHeader.Open()
		if err != nil {
			c.JSON(nethttp.StatusInternalServerError, gin.H{"error": "failed to open uploaded file"})
			return
		}
		defer file.Close()

		// Optional area_id
		var areaID *shareddomain.ID
		if aid := c.PostForm("area_id"); aid != "" {
			id := shareddomain.ID(aid)
			areaID = &id
		}

		cmd := Command{
			WorkspaceID: shareddomain.ID(workspaceID),
			ProjectID:   shareddomain.ID(projectID),
			AreaID:      areaID,
			ItemTypeID:  shareddomain.ID(itemTypeID),
			Filename:    fileHeader.Filename,
			FileSize:    fileHeader.Size,
			FileReader:  file,
			OwnerUserID: shareddomain.ID(userID),
			CreatedBy:   shareddomain.ID(userID),
		}

		result, err := h.Handle(c.Request.Context(), cmd)
		if err != nil {
			c.JSON(nethttp.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(nethttp.StatusCreated, uploadResponse{
			Item: itemResponse{
				ID:            string(result.Item.ID),
				WorkspaceID:   string(result.Item.WorkspaceID),
				ProjectID:     string(result.Item.ProjectID),
				ItemTypeID:    string(result.Item.ItemTypeID),
				LatestVersion: result.Item.LatestVersion,
				Status:        string(result.Item.Status),
				CreatedAt:     result.Item.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			},
			Version: versionResponse{
				ID:      string(result.Version.ID),
				ItemID:  string(result.Version.ItemID),
				Version: result.Version.Version,
				Title:   result.Version.Title,
			},
			Parsing: result.Parsing,
		})
	}
}
