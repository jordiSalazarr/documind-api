package create

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"time"

	"golang.org/x/crypto/bcrypt"

	shared "documind.jordi.org/internal/shared/domain"
)

// Handler handles API key creation.
type Handler struct {
	repo Repository
}

// NewHandler creates a new create API key handler.
func NewHandler(db *sql.DB) *Handler {
	return &Handler{repo: newPostgresRepo(db)}
}

// Handle creates a new API key.
func (h *Handler) Handle(ctx context.Context, cmd Command) (*Result, error) {
	if cmd.Name == "" {
		return nil, ErrInvalidName
	}

	// Generate 32-byte random key → base64url encode
	rawBytes := make([]byte, 32)
	if _, err := rand.Read(rawBytes); err != nil {
		return nil, ErrKeyGenFail
	}
	rawKey := base64.URLEncoding.EncodeToString(rawBytes)

	keyPrefix := rawKey[:8]

	// bcrypt hash the raw key for storage
	keyHash, err := bcrypt.GenerateFromPassword([]byte(rawKey), bcrypt.DefaultCost)
	if err != nil {
		return nil, ErrKeyGenFail
	}

	id := shared.NewID()

	if err := h.repo.Insert(ctx, id.String(), cmd.WorkspaceID, cmd.Name, string(keyHash), keyPrefix, cmd.CreatedBy); err != nil {
		return nil, err
	}

	return &Result{
		ID:        id.String(),
		Name:      cmd.Name,
		KeyPrefix: keyPrefix,
		RawKey:    rawKey,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}, nil
}
