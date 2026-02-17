package register

import "errors"

var (
	ErrInvalidEmail    = errors.New("invalid email address")
	ErrInvalidPassword = errors.New("password must be at least 8 characters")
	ErrInvalidName     = errors.New("name is required")
	ErrEmailTaken      = errors.New("email already registered")
)

type Command struct {
	Email    string
	Password string
	Name     string
}

type Result struct {
	Token string `json:"token"`
	User  UserResult `json:"user"`
}

type UserResult struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}
