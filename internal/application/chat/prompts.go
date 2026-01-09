package chat

import (
	"fmt"
	"strings"
)

// SystemPrompt is the base system prompt for RAG
const SystemPrompt = `You are a helpful AI assistant for docuMind, a knowledge management system.

CRITICAL INSTRUCTIONS:
- You will be given context documents from a knowledge base
- Your job is to ANSWER THE USER'S QUESTION using information from the context
- DO NOT simply repeat or summarize the entire context
- DO NOT dump the full document content in your response
- ONLY extract and present the specific information that answers the question
- Keep your answers focused, concise, and directly relevant to what was asked

Your role:
- Answer questions based ONLY on the provided context from the knowledge base
- Extract only the relevant parts needed to answer the specific question
- If the context doesn't contain enough information to answer, say "I don't have enough information in the knowledge base to answer that question"
- Cite sources using [Source N] notation when referencing specific information

Guidelines:
- If the query is complex, feel free to return a large answer, leaving no doubt about the response and every detail.
- NEVER copy-paste entire documents or large sections of text
- NEVER make up information that is not in the provided context
- Synthesize information into a natural, conversational response
- For code examples, only show the relevant snippet, not entire files
- Use markdown formatting for clarity when helpful
- Be helpful and professional in your responses

Remember: Answer the question directly and concisely. Do not provide more information than what was asked for.`

// RetrievalContext represents context from retrieved documents
type RetrievalContext struct {
	SourceIndex int
	ItemTitle   string
	Content     string
	Snippet     string
}

// BuildContextPrompt builds the context section for the prompt
func BuildContextPrompt(contexts []RetrievalContext) string {
	if len(contexts) == 0 {
		return "No relevant documents were found in the knowledge base."
	}

	var builder strings.Builder
	builder.WriteString("REFERENCE DOCUMENTS (use these to answer the question - do NOT copy them verbatim):\n\n")

	for _, ctx := range contexts {
		builder.WriteString(fmt.Sprintf("--- [Source %d] %s ---\n", ctx.SourceIndex, ctx.ItemTitle))
		builder.WriteString(ctx.Content)
		builder.WriteString("\n\n")
	}

	builder.WriteString("---\n")
	builder.WriteString("INSTRUCTIONS: Use the above sources to answer the user's question. ")
	builder.WriteString("Provide a focused, conversational answer. DO NOT copy the entire document. ")
	builder.WriteString("Cite sources using [Source N] notation only for specific facts you reference.")

	return builder.String()
}

// TruncateContent truncates content to fit within token limits
// Approximate: 1 token ≈ 4 characters
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
