package apikey

import (
	"database/sql"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

// Validator implements middleware.APIKeyChecker by looking up API keys in the database.
type Validator struct {
	db *sql.DB
}

// NewValidator creates a new API key validator.
func NewValidator(db *sql.DB) *Validator {
	return &Validator{db: db}
}

// ValidateKey checks a raw API key against the database.
// It uses the first 8 characters as a prefix to narrow the bcrypt comparison.
func (v *Validator) ValidateKey(rawKey string) (string, error) {
	if len(rawKey) < 8 {
		return "", errors.New("invalid API key")
	}

	prefix := rawKey[:8]

	var id, workspaceID, keyHash string
	err := v.db.QueryRow(`
		SELECT id, workspace_id, key_hash
		FROM api_keys
		WHERE key_prefix = $1
		  AND revoked_at IS NULL
		  AND (expires_at IS NULL OR expires_at > NOW())
	`, prefix).Scan(&id, &workspaceID, &keyHash)

	if errors.Is(err, sql.ErrNoRows) {
		return "", errors.New("invalid API key")
	}
	if err != nil {
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(keyHash), []byte(rawKey)); err != nil {
		return "", errors.New("invalid API key")
	}

	// Async update last_used_at (fire-and-forget)
	go func() {
		_, _ = v.db.Exec(`UPDATE api_keys SET last_used_at = NOW() WHERE id = $1`, id)
	}()

	return workspaceID, nil
}
