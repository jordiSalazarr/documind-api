package domain

import (
	"time"

	shared "documind.jordi.org/internal/shared/domain"
)

type User struct {
	ID           shared.ID
	Email        shared.Email
	PasswordHash string
	Name         string
	shared.Timestamp
}

func NewUser(email shared.Email, passwordHash string, name string) *User {
	return &User{
		ID:           shared.NewID(),
		Email:        email,
		PasswordHash: passwordHash,
		Name:         name,
		Timestamp:    shared.NewTimestamp(),
	}
}

type UserResult struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func ToResult(u *User) *UserResult {
	return &UserResult{
		ID:        string(u.ID),
		Email:     string(u.Email),
		Name:      u.Name,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
