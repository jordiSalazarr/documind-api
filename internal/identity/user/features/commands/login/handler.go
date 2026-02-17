package login

import (
	"context"
	"database/sql"
	"strings"

	"documind.jordi.org/internal/shared/infrastructure/jwt"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	repo       Repository
	jwtService *jwt.Service
}

func NewHandler(db *sql.DB, jwtService *jwt.Service) *Handler {
	return &Handler{
		repo:       newPostgresRepo(db),
		jwtService: jwtService,
	}
}

func (h *Handler) Handle(ctx context.Context, cmd Command) (*Result, error) {
	cmd.Email = strings.TrimSpace(strings.ToLower(cmd.Email))

	if cmd.Email == "" {
		return nil, ErrInvalidEmail
	}
	if cmd.Password == "" {
		return nil, ErrInvalidPassword
	}

	user, err := h.repo.GetByEmail(ctx, cmd.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(cmd.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := h.jwtService.GenerateToken(string(user.ID), string(user.Email), user.Name)
	if err != nil {
		return nil, err
	}

	return &Result{
		Token: token,
		User: UserResult{
			ID:    string(user.ID),
			Email: string(user.Email),
			Name:  user.Name,
		},
	}, nil
}
