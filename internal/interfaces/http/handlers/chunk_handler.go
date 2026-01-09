package handlers

import (
	"context"
	"net/http"

	"documind.jordi.org/internal/domain/common"
	"documind.jordi.org/internal/domain/knowledge"
	"github.com/gin-gonic/gin"
)

type ChunkHandler struct {
	chunkRepo knowledge.ChunkRepository
	itemRepo  knowledge.ItemRepository
}

func NewChunkHandler(chunkRepo knowledge.ChunkRepository, itemRepo knowledge.ItemRepository) *ChunkHandler {
	return &ChunkHandler{
		chunkRepo: chunkRepo,
		itemRepo:  itemRepo,
	}
}

// ChunkDetailResponse represents the full chunk detail for citation viewing
type ChunkDetailResponse struct {
	ID            string                `json:"id"`
	Content       string                `json:"content"`
	Heading       string                `json:"heading"`
	ChunkIndex    int                   `json:"chunk_index"`
	TokenCount    int                   `json:"token_count"`
	CharCount     int                   `json:"char_count"`
	ItemVersionID string                `json:"item_version_id"`
	Metadata      ChunkMetadataResponse `json:"metadata"`
	Item          *ItemInfoResponse     `json:"item,omitempty"`
}

// ChunkMetadataResponse represents chunk metadata in the response
type ChunkMetadataResponse struct {
	StartCharOffset int    `json:"start_char_offset"`
	EndCharOffset   int    `json:"end_char_offset"`
	CodeBlock       bool   `json:"code_block"`
	Language        string `json:"language,omitempty"`
}

// ItemInfoResponse represents basic item information for linking
type ItemInfoResponse struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Summary   string `json:"summary"`
	ProjectID string `json:"project_id"`
	Status    string `json:"status"`
	BodyMd    string `json:"body_md,omitempty"` // Full content for viewing
}

// GetChunk retrieves a chunk by ID with optional item information
func (h *ChunkHandler) GetChunk(c *gin.Context) {
	chunkID := c.Param("id")
	if chunkID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "chunk id is required"})
		return
	}

	chunk, err := h.chunkRepo.GetByID(common.ID(chunkID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "chunk not found"})
		return
	}

	response := ChunkDetailResponse{
		ID:            string(chunk.ID),
		Content:       chunk.Content,
		Heading:       chunk.Heading,
		ChunkIndex:    chunk.ChunkIndex,
		TokenCount:    chunk.TokenCount,
		CharCount:     chunk.CharCount,
		ItemVersionID: string(chunk.ItemVersionID),
		Metadata: ChunkMetadataResponse{
			StartCharOffset: chunk.Metadata.StartCharOffset,
			EndCharOffset:   chunk.Metadata.EndCharOffset,
			CodeBlock:       chunk.Metadata.CodeBlock,
			Language:        chunk.Metadata.Language,
		},
	}

	// Try to get item information
	includeItem := c.Query("include_item") == "true"
	if includeItem && h.itemRepo != nil {
		ctx := context.Background()
		// Get item version first
		itemVersion, err := h.itemRepo.GetItemVersionByID(ctx, chunk.ItemVersionID)
		if err == nil && itemVersion != nil {
			// Get the item itself for project info
			item, err := h.itemRepo.GetItemByID(ctx, itemVersion.ItemID)
			projectID := ""
			status := string(itemVersion.Status)
			if err == nil && item != nil {
				projectID = string(item.ProjectID)
				status = string(item.Status)
			}

			response.Item = &ItemInfoResponse{
				ID:        string(itemVersion.ItemID),
				Title:     itemVersion.Title,
				Summary:   itemVersion.Summary,
				ProjectID: projectID,
				Status:    status,
				BodyMd:    itemVersion.BodyMd,
			}
		}
	}

	c.JSON(http.StatusOK, response)
}

// GetChunksByItem retrieves all chunks for an item version
func (h *ChunkHandler) GetChunksByItem(c *gin.Context) {
	itemVersionID := c.Param("item_version_id")
	if itemVersionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "item_version_id is required"})
		return
	}

	chunks, err := h.chunkRepo.ListByItemVersionID(common.ID(itemVersionID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve chunks"})
		return
	}

	response := make([]ChunkDetailResponse, len(chunks))
	for i, chunk := range chunks {
		response[i] = ChunkDetailResponse{
			ID:            string(chunk.ID),
			Content:       chunk.Content,
			Heading:       chunk.Heading,
			ChunkIndex:    chunk.ChunkIndex,
			TokenCount:    chunk.TokenCount,
			CharCount:     chunk.CharCount,
			ItemVersionID: string(chunk.ItemVersionID),
			Metadata: ChunkMetadataResponse{
				StartCharOffset: chunk.Metadata.StartCharOffset,
				EndCharOffset:   chunk.Metadata.EndCharOffset,
				CodeBlock:       chunk.Metadata.CodeBlock,
				Language:        chunk.Metadata.Language,
			},
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"chunks": response,
		"total":  len(response),
	})
}
