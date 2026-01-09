package knowledge

import (
	"documind.jordi.org/internal/domain/common"
	"errors"
	"time"
)

// SourceType defines the type of evidence
type SourceType string

const (
	SourceTypePDF    SourceType = "pdf"
	SourceTypeURL    SourceType = "url"
	SourceTypeManual SourceType = "manual"
)

func (st SourceType) String() string {
	return string(st)
}

func (st SourceType) IsValid() bool {
	return st == SourceTypePDF || st == SourceTypeURL || st == SourceTypeManual
}

// Source represents evidence (PDFs, URLs, manual entries)
type Source struct {
	ID          common.ID
	WorkspaceID common.ID
	SourceType  SourceType
	Title       string
	URL         *string
	FilePath    *string
	FileSize    *int64
	Date        *time.Time
	Author      *string
	Metadata    map[string]interface{}
	common.Timestamp
	common.Audit
}

func NewSource(
	workspaceID common.ID,
	sourceType SourceType,
	title string,
	createdBy common.ID,
) (*Source, error) {
	if !sourceType.IsValid() {
		return nil, errors.New("invalid source type")
	}

	if title == "" {
		return nil, errors.New("title is required")
	}

	return &Source{
		ID:          common.NewID(),
		WorkspaceID: workspaceID,
		SourceType:  sourceType,
		Title:       title,
		Metadata:    make(map[string]interface{}),
		Timestamp:   common.NewTimestamp(),
		Audit:       common.NewAudit(createdBy),
	}, nil
}

// SetURL sets the URL for URL-type sources
func (s *Source) SetURL(url string) error {
	if s.SourceType != SourceTypeURL {
		return errors.New("can only set URL for URL-type sources")
	}
	s.URL = &url
	return nil
}

// SetFile sets the file information for PDF-type sources
func (s *Source) SetFile(filePath string, fileSize int64) error {
	if s.SourceType != SourceTypePDF {
		return errors.New("can only set file for PDF-type sources")
	}
	s.FilePath = &filePath
	s.FileSize = &fileSize
	return nil
}

// SetDate sets the reference date
func (s *Source) SetDate(date time.Time) {
	s.Date = &date
}

// SetAuthor sets the author
func (s *Source) SetAuthor(author string) {
	s.Author = &author
}

// Update updates the source metadata
func (s *Source) Update(title string, metadata map[string]interface{}, updatedBy common.ID) {
	s.Title = title
	s.Metadata = metadata
	s.Audit.Update(updatedBy)
	s.Timestamp.Update()
}

// ItemSource represents the many-to-many link between items and sources
type ItemSource struct {
	ID          common.ID
	WorkspaceID common.ID
	ItemID      common.ID
	SourceID    common.ID
	LinkedAt    time.Time
	LinkedBy    common.ID
	DeletedAt   *time.Time
}

func NewItemSource(workspaceID, itemID, sourceID, linkedBy common.ID) *ItemSource {
	return &ItemSource{
		ID:          common.NewID(),
		WorkspaceID: workspaceID,
		ItemID:      itemID,
		SourceID:    sourceID,
		LinkedAt:    time.Now(),
		LinkedBy:    linkedBy,
	}
}

// Unlink soft-deletes the item-source link
func (is *ItemSource) Unlink() {
	now := time.Now()
	is.DeletedAt = &now
}

// IsLinked checks if the link is active
func (is *ItemSource) IsLinked() bool {
	return is.DeletedAt == nil
}
