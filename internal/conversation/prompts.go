package conversation

import (
	"fmt"
	"strings"
)

// SystemPrompt is the base system prompt for RAG with few-shot examples and chain-of-thought
const SystemPrompt = `You are a helpful AI assistant for docuMind, a knowledge management system.

RULES:
1. Answer questions based ONLY on the provided context documents.
2. Cite every factual claim using [Source N] notation.
3. NEVER make up information that is not in the provided context.
4. If the context doesn't contain enough information, say so clearly.
5. DO NOT copy-paste entire documents. Synthesize information into focused answers.
6. Use markdown formatting for clarity when helpful.

EXAMPLES:

Example 1 — Factual question with citation:
Q: How do I configure JWT tokens?
A: JWT tokens are configured by setting the ` + "`JWT_SECRET`" + ` environment variable to a minimum 32-character string [Source 1]. The default expiry is 72 hours, configurable via ` + "`JWT_EXPIRY_HOURS`" + ` [Source 1]. You should also ensure the secret is stored securely and rotated periodically [Source 2].

Example 2 — Insufficient context:
Q: What is the pricing for the enterprise plan?
A: I don't have enough information in the provided documents to answer this question. The available sources cover technical setup and configuration but don't include pricing details. You may want to check the pricing page or contact the sales team.

COMPLEX QUERIES:
For complex questions, think step-by-step:
1. Identify the key concepts in the question
2. Find relevant information across all provided sources
3. Synthesize a complete answer with citations for each claim
4. Verify every claim is supported by a source before including it

If the query is complex, provide a thorough answer covering all aspects. Prioritize accuracy over brevity — every statement must be backed by a source.`

// RetrievalContext represents context from retrieved documents
type RetrievalContext struct {
	SourceIndex int
	ItemTitle   string
	Content     string
	Snippet     string
}

// BuildContextPrompt builds the context section for the prompt with clear delimiters
func BuildContextPrompt(contexts []RetrievalContext) string {
	if len(contexts) == 0 {
		return "No relevant documents were found in the knowledge base."
	}

	var builder strings.Builder
	builder.WriteString("=== RETRIEVED DOCUMENTS ===\n\n")

	for _, ctx := range contexts {
		builder.WriteString(fmt.Sprintf("--- [Source %d]: %s ---\n", ctx.SourceIndex, ctx.ItemTitle))
		builder.WriteString(ctx.Content)
		builder.WriteString("\n--- END Source ---\n\n")
	}

	builder.WriteString("=== END DOCUMENTS ===\n\n")
	builder.WriteString("Use the sources above to answer the user's question. ")
	builder.WriteString("Cite each fact with [Source N]. If no source supports a claim, do not include it.")

	return builder.String()
}

// TruncateContent truncates content to fit within token limits
// Approximate: 1 token ~ 4 characters
func TruncateContent(content string, maxTokens int) string {
	maxChars := maxTokens * 4
	if len(content) <= maxChars {
		return content
	}
	return content[:maxChars-3] + "..."
}

// BuildSourceSnippet extracts a relevant snippet from content
func BuildSourceSnippet(content string, maxLength int) string {
	if len(content) <= maxLength {
		return content
	}
	// Try to break at a sentence boundary
	truncated := content[:maxLength]
	lastPeriod := strings.LastIndex(truncated, ". ")
	if lastPeriod > maxLength/2 {
		return truncated[:lastPeriod+1]
	}
	return truncated + "..."
}

// NoContextResponse is returned when no relevant documents are found
const NoContextResponse = `I couldn't find any relevant documents in the knowledge base to answer your question.

You might want to:
1. Try rephrasing your question with different keywords
2. Check if the relevant documentation has been uploaded to this service
3. Ask a more specific question related to the available content`

// SummarySystemPrompt is the system prompt for generating conversation summaries
const SummarySystemPrompt = `You are a conversation summarizer. Your job is to create concise summaries of conversations that capture:
1. The main topics discussed
2. Key questions asked by the user
3. Important information or answers provided
4. Any ongoing context that would be needed to understand follow-up questions

Keep summaries factual and concise. Focus on information that would help understand future questions in the conversation.`

// BuildSummaryPrompt builds the prompt for summary generation
func BuildSummaryPrompt(existingSummary, newConversation string) string {
	if existingSummary == "" {
		return fmt.Sprintf(`Summarize this conversation concisely:

%s

Provide a summary that captures the key topics, questions, and information exchanged.`, newConversation)
	}

	return fmt.Sprintf(`Update this existing conversation summary with new messages:

EXISTING SUMMARY:
%s

NEW MESSAGES TO INCORPORATE:
%s

Provide an updated summary that incorporates the new information while keeping it concise.`, existingSummary, newConversation)
}
