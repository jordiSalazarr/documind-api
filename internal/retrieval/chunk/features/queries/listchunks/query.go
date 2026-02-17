package listchunks

// Query holds input for listing chunks by item version.
type Query struct {
	ItemVersionID string
	Limit         int
	Offset        int
}

// Result is the response payload for a single chunk in the list.
type Result struct {
	ID            string         `json:"id"`
	Content       string         `json:"content"`
	Heading       string         `json:"heading"`
	ChunkIndex    int            `json:"chunk_index"`
	TokenCount    int            `json:"token_count"`
	CharCount     int            `json:"char_count"`
	ItemVersionID string         `json:"item_version_id"`
	Metadata      MetadataResult `json:"metadata"`
}

// MetadataResult holds chunk metadata in the response.
type MetadataResult struct {
	StartCharOffset int    `json:"start_char_offset"`
	EndCharOffset   int    `json:"end_char_offset"`
	CodeBlock       bool   `json:"code_block"`
	Language        string `json:"language,omitempty"`
}
