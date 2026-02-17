package list

// Query contains parameters for listing API keys.
type Query struct {
	WorkspaceID string
}

// Result represents a single API key in the list (never exposes hash or raw key).
type Result struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	KeyPrefix  string  `json:"key_prefix"`
	CreatedBy  string  `json:"created_by"`
	ExpiresAt  *string `json:"expires_at,omitempty"`
	LastUsedAt *string `json:"last_used_at,omitempty"`
	CreatedAt  string  `json:"created_at"`
}
