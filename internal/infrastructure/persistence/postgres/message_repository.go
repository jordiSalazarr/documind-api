package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"documind.jordi.org/internal/domain/chat"
	"documind.jordi.org/internal/domain/common"
)

type MessageRepository struct {
	db *sql.DB
}

func NewMessageRepository(db *sql.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

// Create creates a new message
func (r *MessageRepository) Create(message *chat.Message) error {
	return r.CreateWithContext(context.Background(), message)
}

// CreateWithContext creates a new message with context
func (r *MessageRepository) CreateWithContext(ctx context.Context, message *chat.Message) error {
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

// GetByID retrieves a message by ID
func (r *MessageRepository) GetByID(id common.ID) (*chat.Message, error) {
	return r.GetByIDWithContext(context.Background(), id)
}

// GetByIDWithContext retrieves a message by ID with context
func (r *MessageRepository) GetByIDWithContext(ctx context.Context, id common.ID) (*chat.Message, error) {
	query := `
		SELECT id, conversation_id, workspace_id, role, content, sources,
			token_count, model, latency_ms, created_at
		FROM messages
		WHERE id = $1
	`

	message := &chat.Message{}
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
func (r *MessageRepository) ListByConversationID(conversationID common.ID, limit, offset int) ([]*chat.Message, error) {
	return r.ListByConversationIDWithContext(context.Background(), conversationID, limit, offset)
}

// ListByConversationIDWithContext retrieves messages for a conversation with context
func (r *MessageRepository) ListByConversationIDWithContext(ctx context.Context, conversationID common.ID, limit, offset int) ([]*chat.Message, error) {
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
func (r *MessageRepository) GetLastMessages(conversationID common.ID, limit int) ([]*chat.Message, error) {
	return r.GetLastMessagesWithContext(context.Background(), conversationID, limit)
}

// GetLastMessagesWithContext retrieves the last N messages for context window with context
func (r *MessageRepository) GetLastMessagesWithContext(ctx context.Context, conversationID common.ID, limit int) ([]*chat.Message, error) {
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
func (r *MessageRepository) CountByConversationID(conversationID common.ID) (int, error) {
	return r.CountByConversationIDWithContext(context.Background(), conversationID)
}

// CountByConversationIDWithContext counts messages in a conversation with context
func (r *MessageRepository) CountByConversationIDWithContext(ctx context.Context, conversationID common.ID) (int, error) {
	query := `SELECT COUNT(*) FROM messages WHERE conversation_id = $1`

	var count int
	err := r.db.QueryRowContext(ctx, query, conversationID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count messages: %w", err)
	}

	return count, nil
}

// Helper function to scan multiple messages
func (r *MessageRepository) scanMessages(rows *sql.Rows) ([]*chat.Message, error) {
	messages := []*chat.Message{}
	for rows.Next() {
		message := &chat.Message{}
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

// Helper to convert int to nullable int
func nullableInt(v int) sql.NullInt64 {
	if v == 0 {
		return sql.NullInt64{Valid: false}
	}
	return sql.NullInt64{Int64: int64(v), Valid: true}
}

// Helper to convert string to nullable string
func nullableString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}
