package register

import (
	"context"
	"database/sql"
	"strings"

	userdomain "documind.jordi.org/internal/identity/user/domain"
	shared "documind.jordi.org/internal/shared/domain"
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
	cmd.Name = strings.TrimSpace(cmd.Name)

	if cmd.Email == "" || !strings.Contains(cmd.Email, "@") {
		return nil, ErrInvalidEmail
	}
	if len(cmd.Password) < 8 {
		return nil, ErrInvalidPassword
	}
	if cmd.Name == "" {
		return nil, ErrInvalidName
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(cmd.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	email, err := shared.NewEmail(cmd.Email)
	if err != nil {
		return nil, ErrInvalidEmail
	}

	user := userdomain.NewUser(email, string(hash), cmd.Name)

	if err := h.repo.Insert(ctx, user); err != nil {
		return nil, err
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
