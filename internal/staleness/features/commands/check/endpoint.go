package check

import (
	"errors"
	"net/http"

	"documind.jordi.org/internal/shared/infrastructure/middleware"
	"github.com/gin-gonic/gin"
)

type request struct {
	Diff          string `json:"diff"`
	DiffSummary   string `json:"diff_summary"`
	ProjectID     string `json:"project_id"`
	RepositoryURL string `json:"repository_url"`
	CommitSHA     string `json:"commit_sha"`
}

// Endpoint returns a gin handler for running a staleness check.
func Endpoint(h *Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req request
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{"code": "BAD_REQUEST", "message": "invalid request body"},
			})
			return
		}

		result, err := h.Handle(c.Request.Context(), Command{
			WorkspaceID:   middleware.GetWorkspaceID(c),
			Diff:          req.Diff,
			DiffSummary:   req.DiffSummary,
			ProjectID:     req.ProjectID,
			RepositoryURL: req.RepositoryURL,
			CommitSHA:     req.CommitSHA,
		})
		if err != nil {
			switch {
			case errors.Is(err, ErrEmptyDiff):
				c.JSON(http.StatusBadRequest, gin.H{
					"error": gin.H{"code": "BAD_REQUEST", "message": err.Error()},
				})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": gin.H{"code": "INTERNAL_ERROR", "message": "staleness check failed"},
				})
			}
			return
		}

		c.JSON(http.StatusOK, result)
	}
}
