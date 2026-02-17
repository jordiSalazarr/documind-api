package me

import "errors"

var ErrUserNotFound = errors.New("user not found")

type Query struct {
	UserID string
}

type Result struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}
