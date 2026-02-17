package conversation

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	chatDomain "documind.jordi.org/internal/conversation/domain"
	sharedDomain "documind.jordi.org/internal/shared/domain"
)

// ===== ConversationRepo =====

// ConversationRepo implements both ConversationWriteRepository
// and ConversationReadRepository using PostgreSQL.
type ConversationRepo struct {
	db *sql.DB
}

// NewConversationRepo creates a new PostgreSQL conversation repository
func NewConversationRepo(db *sql.DB) *ConversationRepo {
	return &ConversationRepo{db: db}
}

// --- Write operations ---

// Create creates a new conversation
func (r *ConversationRepo) Create(conversation *chatDomain.Conversation) error {
	return r.CreateWithContext(context.Background(), conversation)
}

// CreateWithContext creates a new conversation with context
func (r *ConversationRepo) CreateWithContext(ctx context.Context, conversation *chatDomain.Conversation) error {
	query := `
		INSERT INTO conversations (
			id, workspace_id, service_id, user_id, title, summary, summary_up_to, message_count, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.db.ExecContext(ctx, query,
		conversation.ID,
		conversation.WorkspaceID,
		conversation.ServiceID,
		conversation.UserID,
		conversation.Title,
		nullableString(conversation.Summary),
		nullableID(conversation.SummaryUpTo),
		conversation.MessageCount,
		conversation.CreatedAt,
		conversation.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create conversation: %w", err)
	}

	return nil
}

// Update updates a conversation
func (r *ConversationRepo) Update(conversation *chatDomain.Conversation) error {
	return r.UpdateWithContext(context.Background(), conversation)
}

// UpdateWithContext updates a conversation with context
func (r *ConversationRepo) UpdateWithContext(ctx context.Context, conversation *chatDomain.Conversation) error {
	query := `
		UPDATE conversations
		SET title = $1, summary = $2, summary_up_to = $3, message_count = $4, updated_at = $5
		WHERE id = $6 AND deleted_at IS NULL
	`

	_, err := r.db.ExecContext(ctx, query,
		conversation.Title,
		nullableString(conversation.Summary),
		nullableID(conversation.SummaryUpTo),
		conversation.MessageCount,
		conversation.UpdatedAt,
		conversation.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update conversation: %w", err)
	}

	return nil
}

// Delete soft-deletes a conversation
func (r *ConversationRepo) Delete(id sharedDomain.ID) error {
	return r.DeleteWithContext(context.Background(), id)
}

// DeleteWithContext soft-deletes a conversation with context
func (r *ConversationRepo) DeleteWithContext(ctx context.Context, id sharedDomain.ID) error {
	query := `UPDATE conversations SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete conversation: %w", err)
	}

	return nil
}

// --- Read operations ---

// GetByID retrieves a conversation by ID
func (r *ConversationRepo) GetByID(id sharedDomain.ID) (*chatDomain.Conversation, error) {
	return r.GetByIDWithContext(context.Background(), id)
}

// GetByIDWithContext retrieves a conversation by ID with context
func (r *ConversationRepo) GetByIDWithContext(ctx context.Context, id sharedDomain.ID) (*chatDomain.Conversation, error) {
	query := `
		SELECT id, workspace_id, service_id, user_id, title, summary, summary_up_to, message_count, created_at, updated_at, deleted_at
		FROM conversations
		WHERE id = $1 AND deleted_at IS NULL
	`

	conversation := &chatDomain.Conversation{}
	var summary sql.NullString
	var summaryUpTo sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&conversation.ID,
		&conversation.WorkspaceID,
		&conversation.ServiceID,
		&conversation.UserID,
		&conversation.Title,
		&summary,
		&summaryUpTo,
		&conversation.MessageCount,
		&conversation.CreatedAt,
		&conversation.UpdatedAt,
		&conversation.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("conversation not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}

	if summary.Valid {
		conversation.Summary = summary.String
	}
	if summaryUpTo.Valid {
		id := sharedDomain.ID(summaryUpTo.String)
		conversation.SummaryUpTo = &id
	}

	return conversation, nil
}

// ListByServiceID retrieves conversations for a service
func (r *ConversationRepo) ListByServiceID(serviceID sharedDomain.ID, limit, offset int) ([]*chatDomain.Conversation, error) {
	return r.ListByServiceIDWithContext(context.Background(), serviceID, limit, offset)
}

// ListByServiceIDWithContext retrieves conversations for a service with context
func (r *ConversationRepo) ListByServiceIDWithContext(ctx context.Context, serviceID sharedDomain.ID, limit, offset int) ([]*chatDomain.Conversation, error) {
	query := `
		SELECT id, workspace_id, service_id, user_id, title, summary, summary_up_to, message_count, created_at, updated_at, deleted_at
		FROM conversations
		WHERE service_id = $1 AND deleted_at IS NULL
		ORDER BY updated_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, serviceID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list conversations: %w", err)
	}
	defer rows.Close()

	return r.scanConversations(rows)
}

// ListByUserID retrieves conversations for a user within a workspace
func (r *ConversationRepo) ListByUserID(workspaceID, userID sharedDomain.ID, limit, offset int) ([]*chatDomain.Conversation, error) {
	return r.ListByUserIDWithContext(context.Background(), workspaceID, userID, limit, offset)
}

// ListByUserIDWithContext retrieves conversations for a user within a workspace with context
func (r *ConversationRepo) ListByUserIDWithContext(ctx context.Context, workspaceID, userID sharedDomain.ID, limit, offset int) ([]*chatDomain.Conversation, error) {
	query := `
		SELECT id, workspace_id, service_id, user_id, title, summary, summary_up_to, message_count, created_at, updated_at, deleted_at
		FROM conversations
		WHERE workspace_id = $1 AND user_id = $2 AND deleted_at IS NULL
		ORDER BY updated_at DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.db.QueryContext(ctx, query, workspaceID, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list conversations: %w", err)
	}
	defer rows.Close()

	return r.scanConversations(rows)
}

// scanConversations scans multiple conversation rows
func (r *ConversationRepo) scanConversations(rows *sql.Rows) ([]*chatDomain.Conversation, error) {
	conversations := []*chatDomain.Conversation{}
	for rows.Next() {
		conversation := &chatDomain.Conversation{}
		var summary sql.NullString
		var summaryUpTo sql.NullString
		err := rows.Scan(
			&conversation.ID,
			&conversation.WorkspaceID,
			&conversation.ServiceID,
			&conversation.UserID,
			&conversation.Title,
			&summary,
			&summaryUpTo,
			&conversation.MessageCount,
			&conversation.CreatedAt,
			&conversation.UpdatedAt,
			&conversation.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan conversation: %w", err)
		}

		if summary.Valid {
			conversation.Summary = summary.String
		}
		if summaryUpTo.Valid {
			id := sharedDomain.ID(summaryUpTo.String)
			conversation.SummaryUpTo = &id
		}

		conversations = append(conversations, conversation)
	}

	return conversations, nil
}

// ===== MessageRepo =====

// MessageRepo implements both MessageWriteRepository
// and MessageReadRepository using PostgreSQL.
type MessageRepo struct {
	db *sql.DB
}

// NewMessageRepo creates a new PostgreSQL message repository
func NewMessageRepo(db *sql.DB) *MessageRepo {
	return &MessageRepo{db: db}
}

// --- Write operations ---

// Create creates a new message
func (r *MessageRepo) Create(message *chatDomain.Message) error {
	return r.CreateWithContext(context.Background(), message)
}

// CreateWithContext creates a new message with context
func (r *MessageRepo) CreateWithContext(ctx context.Context, message *chatDomain.Message) error {
	sourcesJSON, err := json.Marshal(message.Sources)
	if err != nil {
		return fmt.Errorf("failed to marshal sources: %w", err)
	}

	query := `
		INSERT INTO messages (
			id, conversation_id, workspace_id, role, content, sources,
			token_count, model, latency_ms, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err = r.db.ExecContext(ctx, query,
		message.ID,
		message.ConversationID,
		message.WorkspaceID,
		message.Role,
		message.Content,
		sourcesJSON,
		nullableInt(message.TokenCount),
		nullableString(message.Model),
		nullableInt(message.LatencyMs),
		message.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}

	return nil
}

// --- Read operations ---

// GetByID retrieves a message by ID
func (r *MessageRepo) GetByID(id sharedDomain.ID) (*chatDomain.Message, error) {
	return r.GetByIDWithContext(context.Background(), id)
}

// GetByIDWithContext retrieves a message by ID with context
func (r *MessageRepo) GetByIDWithContext(ctx context.Context, id sharedDomain.ID) (*chatDomain.Message, error) {
	query := `
		SELECT id, conversation_id, workspace_id, role, content, sources,
			token_count, model, latency_ms, created_at
		FROM messages
		WHERE id = $1
	`

	message := &chatDomain.Message{}
	var sourcesJSON []byte
	var tokenCount sql.NullInt64
	var model sql.NullString
	var latencyMs sql.NullInt64

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&message.ID,
		&message.ConversationID,
		&message.WorkspaceID,
		&message.Role,
		&message.Content,
		&sourcesJSON,
		&tokenCount,
		&model,
		&latencyMs,
		&message.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("message not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	if err := json.Unmarshal(sourcesJSON, &message.Sources); err != nil {
		return nil, fmt.Errorf("failed to unmarshal sources: %w", err)
	}

	if tokenCount.Valid {
		message.TokenCount = int(tokenCount.Int64)
	}
	if model.Valid {
		message.Model = model.String
	}
	if latencyMs.Valid {
		message.LatencyMs = int(latencyMs.Int64)
	}

	return message, nil
}

// ListByConversationID retrieves messages for a conversation (ordered by created_at)
func (r *MessageRepo) ListByConversationID(conversationID sharedDomain.ID, limit, offset int) ([]*chatDomain.Message, error) {
	return r.ListByConversationIDWithContext(context.Background(), conversationID, limit, offset)
}

// ListByConversationIDWithContext retrieves messages for a conversation with context
func (r *MessageRepo) ListByConversationIDWithContext(ctx context.Context, conversationID sharedDomain.ID, limit, offset int) ([]*chatDomain.Message, error) {
	query := `
		SELECT id, conversation_id, workspace_id, role, content, sources,
			token_count, model, latency_ms, created_at
		FROM messages
		WHERE conversation_id = $1
		ORDER BY created_at ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, conversationID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list messages: %w", err)
	}
	defer rows.Close()

	return r.scanMessages(rows)
}

// GetLastMessages retrieves the last N messages for context window
func (r *MessageRepo) GetLastMessages(conversationID sharedDomain.ID, limit int) ([]*chatDomain.Message, error) {
	return r.GetLastMessagesWithContext(context.Background(), conversationID, limit)
}

// GetLastMessagesWithContext retrieves the last N messages for context window with context
func (r *MessageRepo) GetLastMessagesWithContext(ctx context.Context, conversationID sharedDomain.ID, limit int) ([]*chatDomain.Message, error) {
	// Subquery to get last N messages, then order them ASC
	query := `
		SELECT * FROM (
			SELECT id, conversation_id, workspace_id, role, content, sources,
				token_count, model, latency_ms, created_at
			FROM messages
			WHERE conversation_id = $1
			ORDER BY created_at DESC
			LIMIT $2
		) subquery
		ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, conversationID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get last messages: %w", err)
	}
	defer rows.Close()

	return r.scanMessages(rows)
}

// CountByConversationID counts messages in a conversation
func (r *MessageRepo) CountByConversationID(conversationID sharedDomain.ID) (int, error) {
	return r.CountByConversationIDWithContext(context.Background(), conversationID)
}

// CountByConversationIDWithContext counts messages in a conversation with context
func (r *MessageRepo) CountByConversationIDWithContext(ctx context.Context, conversationID sharedDomain.ID) (int, error) {
	query := `SELECT COUNT(*) FROM messages WHERE conversation_id = $1`

	var count int
	err := r.db.QueryRowContext(ctx, query, conversationID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count messages: %w", err)
	}

	return count, nil
}

// scanMessages scans multiple message rows
func (r *MessageRepo) scanMessages(rows *sql.Rows) ([]*chatDomain.Message, error) {
	messages := []*chatDomain.Message{}
	for rows.Next() {
		message := &chatDomain.Message{}
		var sourcesJSON []byte
		var tokenCount sql.NullInt64
		var model sql.NullString
		var latencyMs sql.NullInt64

		err := rows.Scan(
			&message.ID,
			&message.ConversationID,
			&message.WorkspaceID,
			&message.Role,
			&message.Content,
			&sourcesJSON,
			&tokenCount,
			&model,
			&latencyMs,
			&message.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}

		if err := json.Unmarshal(sourcesJSON, &message.Sources); err != nil {
			return nil, fmt.Errorf("failed to unmarshal sources: %w", err)
		}

		if tokenCount.Valid {
			message.TokenCount = int(tokenCount.Int64)
		}
		if model.Valid {
			message.Model = model.String
		}
		if latencyMs.Valid {
			message.LatencyMs = int(latencyMs.Int64)
		}

		messages = append(messages, message)
	}

	return messages, nil
}

// ===== Helper functions =====

// nullableID converts a pointer to sharedDomain.ID to sql.NullString
func nullableID(id *sharedDomain.ID) sql.NullString {
	if id == nil {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: string(*id), Valid: true}
}

// nullableString converts a string to sql.NullString
func nullableString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

// nullableInt converts int to sql.NullInt64
func nullableInt(v int) sql.NullInt64 {
	if v == 0 {
		return sql.NullInt64{Valid: false}
	}
	return sql.NullInt64{Int64: int64(v), Valid: true}
}
