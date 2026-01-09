package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"documind.jordi.org/internal/domain/chat"
	"documind.jordi.org/internal/domain/common"
)

type ConversationRepository struct {
	db *sql.DB
}

func NewConversationRepository(db *sql.DB) *ConversationRepository {
	return &ConversationRepository{db: db}
}

// Create creates a new conversation
func (r *ConversationRepository) Create(conversation *chat.Conversation) error {
	return r.CreateWithContext(context.Background(), conversation)
}

// CreateWithContext creates a new conversation with context
func (r *ConversationRepository) CreateWithContext(ctx context.Context, conversation *chat.Conversation) error {
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

// nullableID converts a pointer to common.ID to sql.NullString
func nullableID(id *common.ID) sql.NullString {
	if id == nil {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: string(*id), Valid: true}
}

// GetByID retrieves a conversation by ID
func (r *ConversationRepository) GetByID(id common.ID) (*chat.Conversation, error) {
	return r.GetByIDWithContext(context.Background(), id)
}

// GetByIDWithContext retrieves a conversation by ID with context
func (r *ConversationRepository) GetByIDWithContext(ctx context.Context, id common.ID) (*chat.Conversation, error) {
	query := `
		SELECT id, workspace_id, service_id, user_id, title, summary, summary_up_to, message_count, created_at, updated_at, deleted_at
		FROM conversations
		WHERE id = $1 AND deleted_at IS NULL
	`

	conversation := &chat.Conversation{}
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
		id := common.ID(summaryUpTo.String)
		conversation.SummaryUpTo = &id
	}

	return conversation, nil
}

// ListByServiceID retrieves conversations for a service
func (r *ConversationRepository) ListByServiceID(serviceID common.ID, limit, offset int) ([]*chat.Conversation, error) {
	return r.ListByServiceIDWithContext(context.Background(), serviceID, limit, offset)
}

// ListByServiceIDWithContext retrieves conversations for a service with context
func (r *ConversationRepository) ListByServiceIDWithContext(ctx context.Context, serviceID common.ID, limit, offset int) ([]*chat.Conversation, error) {
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
func (r *ConversationRepository) ListByUserID(workspaceID, userID common.ID, limit, offset int) ([]*chat.Conversation, error) {
	return r.ListByUserIDWithContext(context.Background(), workspaceID, userID, limit, offset)
}

// ListByUserIDWithContext retrieves conversations for a user within a workspace with context
func (r *ConversationRepository) ListByUserIDWithContext(ctx context.Context, workspaceID, userID common.ID, limit, offset int) ([]*chat.Conversation, error) {
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
func (r *ConversationRepository) scanConversations(rows *sql.Rows) ([]*chat.Conversation, error) {
	conversations := []*chat.Conversation{}
	for rows.Next() {
		conversation := &chat.Conversation{}
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
			id := common.ID(summaryUpTo.String)
			conversation.SummaryUpTo = &id
		}

		conversations = append(conversations, conversation)
	}

	return conversations, nil
}

// Update updates a conversation
func (r *ConversationRepository) Update(conversation *chat.Conversation) error {
	return r.UpdateWithContext(context.Background(), conversation)
}

// UpdateWithContext updates a conversation with context
func (r *ConversationRepository) UpdateWithContext(ctx context.Context, conversation *chat.Conversation) error {
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
func (r *ConversationRepository) Delete(id common.ID) error {
	return r.DeleteWithContext(context.Background(), id)
}

// DeleteWithContext soft-deletes a conversation with context
func (r *ConversationRepository) DeleteWithContext(ctx context.Context, id common.ID) error {
	query := `UPDATE conversations SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete conversation: %w", err)
	}

	return nil
}
