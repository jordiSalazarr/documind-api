package check

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"documind.jordi.org/internal/shared/infrastructure/openai"
	"documind.jordi.org/internal/staleness/domain"

	shared "documind.jordi.org/internal/shared/domain"
)

// Handler handles the staleness check pipeline.
type Handler struct {
	repo       Repository
	retriever  Retriever
	chatClient ChatClient
}

// NewHandler creates a new staleness check handler.
func NewHandler(db *sql.DB, retriever Retriever, chatClient ChatClient) *Handler {
	return &Handler{
		repo:       newPostgresRepo(db),
		retriever:  retriever,
		chatClient: chatClient,
	}
}

// Handle runs the 5-step staleness detection pipeline.
func (h *Handler) Handle(ctx context.Context, cmd Command) (*Result, error) {
	if cmd.Diff == "" && cmd.DiffSummary == "" {
		return nil, ErrEmptyDiff
	}

	// Step 1: Summarize diff if needed
	diffSummary := cmd.DiffSummary
	if diffSummary == "" {
		diffSummary = h.summarizeDiff(ctx, cmd.Diff)
	}

	// Step 2: Retrieve related docs
	results, err := h.retriever.Retrieve(ctx, RetrievalInput{
		Query:       diffSummary,
		WorkspaceID: cmd.WorkspaceID,
		ProjectID:   cmd.ProjectID,
		MaxResults:  10,
	})
	if err != nil {
		return nil, fmt.Errorf("retrieval failed: %w", err)
	}

	// Step 3: Group by item — deduplicate chunks by ItemID
	grouped := h.groupByItem(results)

	// Step 4: Evaluate staleness for each unique item
	itemResults := h.evaluateItems(ctx, diffSummary, grouped)

	// Step 5: Persist report
	var projectID *shared.ID
	if cmd.ProjectID != "" {
		pid := shared.ID(cmd.ProjectID)
		projectID = &pid
	}

	report := domain.NewReport(
		shared.ID(cmd.WorkspaceID),
		projectID,
		domain.TriggerCodeChange,
		cmd.CommitSHA,
		cmd.RepositoryURL,
		diffSummary,
	)

	if err := h.repo.InsertReport(ctx, report); err != nil {
		return nil, fmt.Errorf("failed to persist report: %w", err)
	}

	var reportItems []*domain.StalenessReportItem
	staleCount := 0
	for _, ir := range itemResults {
		reportItems = append(reportItems, domain.NewReportItem(
			report.ID,
			shared.ID(ir.ItemID),
			shared.ID(ir.ItemVersionID),
			ir.ItemTitle,
			ir.Stale,
			ir.Confidence,
			ir.Explanation,
			ir.SuggestedUpdate,
		))
		if ir.Stale {
			staleCount++
		}
	}

	if err := h.repo.InsertReportItems(ctx, reportItems); err != nil {
		return nil, fmt.Errorf("failed to persist report items: %w", err)
	}

	return &Result{
		ReportID:   report.ID.String(),
		StaleCount: staleCount,
		Total:      len(itemResults),
		Items:      itemResults,
	}, nil
}

// summarizeDiff summarizes a long diff using LLM, or truncates as fallback.
func (h *Handler) summarizeDiff(ctx context.Context, diff string) string {
	if len(diff) <= 4000 {
		return diff
	}

	resp, err := h.chatClient.CreateCompletion(ctx, []openai.ChatMessage{
		{Role: "system", Content: "You are a concise code change summarizer. Summarize the following code diff in 2-3 sentences, focusing on what functionality changed and which areas of the codebase were affected."},
		{Role: "user", Content: diff[:8000]}, // Cap input to avoid token overflow
	}, 300)
	if err != nil {
		log.Printf("Failed to summarize diff via LLM, truncating: %v", err)
		if len(diff) > 500 {
			return diff[:500]
		}
		return diff
	}

	if len(resp.Choices) > 0 {
		return resp.Choices[0].Message.Content
	}

	if len(diff) > 500 {
		return diff[:500]
	}
	return diff
}

// groupedItem holds aggregated content for a unique item.
type groupedItem struct {
	ItemID        string
	ItemVersionID string
	ItemTitle     string
	Content       string
}

// groupByItem deduplicates retrieval results by ItemID and concatenates content.
func (h *Handler) groupByItem(results []*RetrievalResult) []*groupedItem {
	seen := map[string]*groupedItem{}
	var order []string

	for _, r := range results {
		if r.ItemID == "" {
			continue
		}
		if existing, ok := seen[r.ItemID]; ok {
			// Append content up to 3000 chars total
			if len(existing.Content) < 3000 {
				remaining := 3000 - len(existing.Content)
				addition := r.Content
				if len(addition) > remaining {
					addition = addition[:remaining]
				}
				existing.Content += "\n\n" + addition
			}
		} else {
			content := r.Content
			if len(content) > 3000 {
				content = content[:3000]
			}
			seen[r.ItemID] = &groupedItem{
				ItemID:        r.ItemID,
				ItemVersionID: r.ItemVersionID,
				ItemTitle:     r.ItemTitle,
				Content:       content,
			}
			order = append(order, r.ItemID)
		}
	}

	grouped := make([]*groupedItem, 0, len(order))
	for _, id := range order {
		grouped = append(grouped, seen[id])
	}
	return grouped
}

// stalenessEval represents the structured JSON response from the LLM.
type stalenessEval struct {
	Stale           bool    `json:"stale"`
	Confidence      float64 `json:"confidence"`
	Explanation     string  `json:"explanation"`
	SuggestedUpdate string  `json:"suggested_update"`
}

// evaluateItems evaluates each grouped item for staleness using the LLM.
func (h *Handler) evaluateItems(ctx context.Context, diffSummary string, items []*groupedItem) []*ItemResult {
	results := make([]*ItemResult, 0, len(items))

	for _, item := range items {
		ir := h.evaluateOneItem(ctx, diffSummary, item)
		results = append(results, ir)
	}

	return results
}

func (h *Handler) evaluateOneItem(ctx context.Context, diffSummary string, item *groupedItem) *ItemResult {
	prompt := fmt.Sprintf(`You are a documentation staleness detector. Given a code change summary and a documentation excerpt, determine if the documentation is likely stale (outdated or inaccurate due to the code change).

Respond with ONLY a JSON object (no markdown, no code fences):
{"stale": true/false, "confidence": 0.0-1.0, "explanation": "brief reason", "suggested_update": "what should change (empty string if not stale)"}

=== CODE CHANGE SUMMARY ===
%s

=== DOCUMENT: %s ===
%s`, diffSummary, item.ItemTitle, item.Content)

	resp, err := h.chatClient.CreateCompletion(ctx, []openai.ChatMessage{
		{Role: "user", Content: prompt},
	}, 400)
	if err != nil {
		log.Printf("Failed to evaluate staleness for item %s: %v", item.ItemID, err)
		return &ItemResult{
			ItemID:        item.ItemID,
			ItemVersionID: item.ItemVersionID,
			ItemTitle:     item.ItemTitle,
			Stale:         false,
			Confidence:    0.0,
			Explanation:   fmt.Sprintf("evaluation failed: %v", err),
		}
	}

	if len(resp.Choices) == 0 {
		return &ItemResult{
			ItemID:        item.ItemID,
			ItemVersionID: item.ItemVersionID,
			ItemTitle:     item.ItemTitle,
			Stale:         false,
			Confidence:    0.0,
			Explanation:   "no response from evaluator",
		}
	}

	content := strings.TrimSpace(resp.Choices[0].Message.Content)
	// Strip markdown code fences if present
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var eval stalenessEval
	if err := json.Unmarshal([]byte(content), &eval); err != nil {
		log.Printf("Failed to parse staleness eval for item %s: %v (raw: %s)", item.ItemID, err, content)
		return &ItemResult{
			ItemID:        item.ItemID,
			ItemVersionID: item.ItemVersionID,
			ItemTitle:     item.ItemTitle,
			Stale:         false,
			Confidence:    0.0,
			Explanation:   fmt.Sprintf("failed to parse evaluation: %v", err),
		}
	}

	return &ItemResult{
		ItemID:          item.ItemID,
		ItemVersionID:   item.ItemVersionID,
		ItemTitle:       item.ItemTitle,
		Stale:           eval.Stale,
		Confidence:      eval.Confidence,
		Explanation:     eval.Explanation,
		SuggestedUpdate: eval.SuggestedUpdate,
	}
}
