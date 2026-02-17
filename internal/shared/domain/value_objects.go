package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// ID represents a domain entity identifier
type ID string

func NewID() ID {
	return ID(uuid.New().String())
}

func (id ID) String() string {
	return string(id)
}

func (id ID) IsEmpty() bool {
	return string(id) == ""
}

// Timestamp represents domain timestamps
type Timestamp struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

func NewTimestamp() Timestamp {
	now := time.Now()
	return Timestamp{
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (t *Timestamp) Update() {
	t.UpdatedAt = time.Now()
}

func (t *Timestamp) SoftDelete() {
	now := time.Now()
	t.DeletedAt = &now
}

func (t *Timestamp) IsDeleted() bool {
	return t.DeletedAt != nil
}

// Audit represents audit information
type Audit struct {
	CreatedBy ID
	UpdatedBy ID
}

func NewAudit(userID ID) Audit {
	return Audit{
		CreatedBy: userID,
		UpdatedBy: userID,
	}
}

func (a *Audit) Update(userID ID) {
	a.UpdatedBy = userID
}

// Email value object
type Email string

func NewEmail(email string) (Email, error) {
	if email == "" {
		return "", errors.New("email cannot be empty")
	}
	// Basic validation - in production, use a proper email validation library
	if len(email) < 3 || len(email) > 255 {
		return "", errors.New("email length must be between 3 and 255 characters")
	}
	return Email(email), nil
}

func (e Email) String() string {
	return string(e)
}

// Slug value object
type Slug string

func NewSlug(s string) (Slug, error) {
	if s == "" {
		return "", errors.New("slug cannot be empty")
	}
	// In production, implement proper slug validation (lowercase, hyphens, etc.)
	return Slug(s), nil
}

func (s Slug) String() string {
	return string(s)
}
