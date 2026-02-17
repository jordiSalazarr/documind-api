package create

import "errors"

var (
	ErrInvalidName = errors.New("API key name is required")
	ErrKeyGenFail  = errors.New("failed to generate API key")
)

// Command contains parameters for creating an API key.
type Command struct {
	WorkspaceID string
	Name        string
	CreatedBy   string
}

// Result contains the created API key details.
// RawKey is only returned once at creation time — it is never stored.
type Result struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	KeyPrefix string `json:"key_prefix"`
	RawKey    string `json:"raw_key"`
	CreatedAt string `json:"created_at"`
}
