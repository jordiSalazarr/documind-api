package handlers

import (
	"net/http"
	"strconv"

	"documind.jordi.org/internal/application/item"
	"documind.jordi.org/internal/domain/common"
	"documind.jordi.org/internal/domain/knowledge"
	"github.com/gin-gonic/gin"
)

type ItemHandler struct {
	service *item.Service
}

func NewItemHandler(service *item.Service) *ItemHandler {
	return &ItemHandler{service: service}
}

// Request/Response structs

type CreateItemRequest struct {
	WorkspaceID  string                 `json:"workspace_id" binding:"required"`
	ProjectID    string                 `json:"project_id" binding:"required"`
	AreaID       *string                `json:"area_id"`
	ItemTypeID   string                 `json:"item_type_id" binding:"required"`
	Title        string                 `json:"title" binding:"required"`
	Summary      string                 `json:"summary"`
	BodyMd       string                 `json:"body_md"`
	CustomFields map[string]interface{} `json:"custom_fields"`
	Tags         []string               `json:"tags"`
	OwnerUserID  string                 `json:"owner_user_id" binding:"required"`
	CreatedBy    string                 `json:"created_by" binding:"required"`
}

type CreateItemVersionRequest struct {
	Title        string                 `json:"title" binding:"required"`
	Summary      string                 `json:"summary"`
	BodyMd       string                 `json:"body_md"`
	CustomFields map[string]interface{} `json:"custom_fields"`
	Tags         []string               `json:"tags"`
	CreatedBy    string                 `json:"created_by" binding:"required"`
}

type UpdateItemRequest struct {
	ProjectID   *string `json:"project_id"`
	AreaID      *string `json:"area_id"`
	OwnerUserID *string `json:"owner_user_id"`
	UpdatedBy   string  `json:"updated_by" binding:"required"`
}

type ItemResponse struct {
	ID            string  `json:"id"`
	WorkspaceID   string  `json:"workspace_id"`
	ProjectID     string  `json:"project_id"`
	AreaID        *string `json:"area_id"`
	ItemTypeID    string  `json:"item_type_id"`
	LatestVersion int     `json:"latest_version"`
	Status        string  `json:"status"`
	OwnerUserID   string  `json:"owner_user_id"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
	CreatedBy     string  `json:"created_by"`
	UpdatedBy     string  `json:"updated_by"`
}

// ItemListResponse includes title/summary from the latest version for display in lists
type ItemListResponse struct {
	ID            string   `json:"id"`
	WorkspaceID   string   `json:"workspace_id"`
	ProjectID     string   `json:"project_id"`
	AreaID        *string  `json:"area_id"`
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

type ItemWithVersionResponse struct {
	Item    ItemResponse        `json:"item"`
	Version ItemVersionResponse `json:"version"`
}

// ===== Item Handlers =====

func (h *ItemHandler) Create(c *gin.Context) {
	var req CreateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var areaID *common.ID
	if req.AreaID != nil {
		id := common.ID(*req.AreaID)
		areaID = &id
	}

	input := item.CreateItemInput{
		WorkspaceID:  common.ID(req.WorkspaceID),
		ProjectID:    common.ID(req.ProjectID),
		AreaID:       areaID,
		ItemTypeID:   common.ID(req.ItemTypeID),
		Title:        req.Title,
		Summary:      req.Summary,
		BodyMd:       req.BodyMd,
		CustomFields: req.CustomFields,
		Tags:         req.Tags,
		OwnerUserID:  common.ID(req.OwnerUserID),
		CreatedBy:    common.ID(req.CreatedBy),
	}

	createdItem, version, err := h.service.CreateItem(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var areaIDStr *string
	if createdItem.ServiceID != nil {
		s := string(*createdItem.ServiceID)
		areaIDStr = &s
	}

	c.JSON(http.StatusCreated, ItemWithVersionResponse{
		Item: ItemResponse{
			ID:            string(createdItem.ID),
			WorkspaceID:   string(createdItem.WorkspaceID),
			ProjectID:     string(createdItem.ProjectID),
			AreaID:        areaIDStr,
			ItemTypeID:    string(createdItem.ItemTypeID),
			LatestVersion: createdItem.LatestVersion,
			Status:        string(createdItem.Status),
			OwnerUserID:   string(createdItem.OwnerUserID),
			CreatedAt:     createdItem.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:     createdItem.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
			CreatedBy:     string(createdItem.CreatedBy),
			UpdatedBy:     string(createdItem.UpdatedBy),
		},
		Version: ItemVersionResponse{
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
		},
	})
}

func (h *ItemHandler) Get(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	item, version, err := h.service.GetItemWithLatestVersion(c.Request.Context(), common.ID(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	var areaIDStr *string
	if item.ServiceID != nil {
		s := string(*item.ServiceID)
		areaIDStr = &s
	}

	c.JSON(http.StatusOK, ItemWithVersionResponse{
		Item: ItemResponse{
			ID:            string(item.ID),
			WorkspaceID:   string(item.WorkspaceID),
			ProjectID:     string(item.ProjectID),
			AreaID:        areaIDStr,
			ItemTypeID:    string(item.ItemTypeID),
			LatestVersion: item.LatestVersion,
			Status:        string(item.Status),
			OwnerUserID:   string(item.OwnerUserID),
			CreatedAt:     item.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:     item.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
			CreatedBy:     string(item.CreatedBy),
			UpdatedBy:     string(item.UpdatedBy),
		},
		Version: ItemVersionResponse{
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
		},
	})
}

func (h *ItemHandler) ListByProject(c *gin.Context) {
	projectID := c.Query("project_id")
	serviceID := c.Query("service_id")

	// Support both project_id and service_id
	if projectID == "" && serviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "either project_id or service_id is required"})
		return
	}

	var items []*knowledge.Item
	var err error

	if serviceID != "" {
		items, err = h.service.ListItemsByArea(c.Request.Context(), common.ID(serviceID))
	} else {
		items, err = h.service.ListItemsByProject(c.Request.Context(), common.ID(projectID))
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var response []ItemListResponse
	for _, item := range items {
		var areaIDStr *string
		if item.ServiceID != nil {
			s := string(*item.ServiceID)
			areaIDStr = &s
		}

		// Fetch the latest version to get title, summary, and tags
		var title, summary string
		var tags []string
		version, verErr := h.service.GetLatestItemVersion(c.Request.Context(), item.ID)
		if verErr == nil && version != nil {
			title = version.Title
			summary = version.Summary
			tags = version.Tags
		}

		response = append(response, ItemListResponse{
			ID:            string(item.ID),
			WorkspaceID:   string(item.WorkspaceID),
			ProjectID:     string(item.ProjectID),
			AreaID:        areaIDStr,
			ItemTypeID:    string(item.ItemTypeID),
			Title:         title,
			Summary:       summary,
			Tags:          tags,
			LatestVersion: item.LatestVersion,
			Status:        string(item.Status),
			OwnerUserID:   string(item.OwnerUserID),
			CreatedAt:     item.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:     item.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
			CreatedBy:     string(item.CreatedBy),
			UpdatedBy:     string(item.UpdatedBy),
		})
	}

	c.JSON(http.StatusOK, response)
}

func (h *ItemHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	var req UpdateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := item.UpdateItemInput{
		ID:        common.ID(id),
		UpdatedBy: common.ID(req.UpdatedBy),
	}

	if req.ProjectID != nil {
		pid := common.ID(*req.ProjectID)
		input.ProjectID = &pid
	}

	if req.AreaID != nil {
		aid := common.ID(*req.AreaID)
		input.AreaID = &aid
	}

	if req.OwnerUserID != nil {
		oid := common.ID(*req.OwnerUserID)
		input.OwnerUserID = &oid
	}

	updatedItem, err := h.service.UpdateItem(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var areaIDStr *string
	if updatedItem.ServiceID != nil {
		s := string(*updatedItem.ServiceID)
		areaIDStr = &s
	}

	c.JSON(http.StatusOK, ItemResponse{
		ID:            string(updatedItem.ID),
		WorkspaceID:   string(updatedItem.WorkspaceID),
		ProjectID:     string(updatedItem.ProjectID),
		AreaID:        areaIDStr,
		ItemTypeID:    string(updatedItem.ItemTypeID),
		LatestVersion: updatedItem.LatestVersion,
		Status:        string(updatedItem.Status),
		OwnerUserID:   string(updatedItem.OwnerUserID),
		CreatedAt:     updatedItem.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     updatedItem.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		CreatedBy:     string(updatedItem.CreatedBy),
		UpdatedBy:     string(updatedItem.UpdatedBy),
	})
}

func (h *ItemHandler) Publish(c *gin.Context) {
	id := c.Param("id")
	updatedBy := c.Query("updated_by")

	if id == "" || updatedBy == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id and updated_by are required"})
		return
	}

	updatedItem, err := h.service.PublishItem(c.Request.Context(), common.ID(id), common.ID(updatedBy))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var areaIDStr *string
	if updatedItem.ServiceID != nil {
		s := string(*updatedItem.ServiceID)
		areaIDStr = &s
	}

	c.JSON(http.StatusOK, ItemResponse{
		ID:            string(updatedItem.ID),
		WorkspaceID:   string(updatedItem.WorkspaceID),
		ProjectID:     string(updatedItem.ProjectID),
		AreaID:        areaIDStr,
		ItemTypeID:    string(updatedItem.ItemTypeID),
		LatestVersion: updatedItem.LatestVersion,
		Status:        string(updatedItem.Status),
		OwnerUserID:   string(updatedItem.OwnerUserID),
		CreatedAt:     updatedItem.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     updatedItem.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		CreatedBy:     string(updatedItem.CreatedBy),
		UpdatedBy:     string(updatedItem.UpdatedBy),
	})
}

func (h *ItemHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	if err := h.service.DeleteItem(c.Request.Context(), common.ID(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// ===== ItemVersion Handlers =====

func (h *ItemHandler) CreateVersion(c *gin.Context) {
	itemID := c.Param("id")
	if itemID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "item_id is required"})
		return
	}

	var req CreateItemVersionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get workspace ID from the item
	existingItem, err := h.service.GetItem(c.Request.Context(), common.ID(itemID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
		return
	}

	input := item.CreateItemVersionInput{
		ItemID:       common.ID(itemID),
		WorkspaceID:  existingItem.WorkspaceID,
		Title:        req.Title,
		Summary:      req.Summary,
		BodyMd:       req.BodyMd,
		CustomFields: req.CustomFields,
		Tags:         req.Tags,
		CreatedBy:    common.ID(req.CreatedBy),
	}

	version, err := h.service.CreateItemVersion(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, ItemVersionResponse{
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

func (h *ItemHandler) GetVersion(c *gin.Context) {
	itemID := c.Param("id")
	versionStr := c.Param("version")

	if itemID == "" || versionStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "item_id and version are required"})
		return
	}

	version, err := strconv.Atoi(versionStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid version number"})
		return
	}

	itemVersion, err := h.service.GetItemVersion(c.Request.Context(), common.ID(itemID), version)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ItemVersionResponse{
		ID:           string(itemVersion.ID),
		ItemID:       string(itemVersion.ItemID),
		WorkspaceID:  string(itemVersion.WorkspaceID),
		Version:      itemVersion.Version,
		Title:        itemVersion.Title,
		Summary:      itemVersion.Summary,
		BodyMd:       itemVersion.BodyMd,
		CustomFields: itemVersion.CustomFields,
		Tags:         itemVersion.Tags,
		Status:       string(itemVersion.Status),
		CreatedAt:    itemVersion.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		CreatedBy:    string(itemVersion.CreatedBy),
	})
}

func (h *ItemHandler) ListVersions(c *gin.Context) {
	itemID := c.Param("id")
	if itemID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "item_id is required"})
		return
	}

	versions, err := h.service.ListItemVersions(c.Request.Context(), common.ID(itemID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var response []ItemVersionResponse
	for _, version := range versions {
		response = append(response, ItemVersionResponse{
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

	c.JSON(http.StatusOK, response)
}
