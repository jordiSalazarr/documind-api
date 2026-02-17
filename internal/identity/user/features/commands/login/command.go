package login

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidEmail       = errors.New("email is required")
	ErrInvalidPassword    = errors.New("password is required")
)

type Command struct {
	Email    string
	Password string
}

type Result struct {
	Token string     `json:"token"`
	User  UserResult `json:"user"`
}

type UserResult struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}
