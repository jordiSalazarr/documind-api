package workspace

import (
	"context"
	"errors"
	"strings"

	"documind.jordi.org/internal/domain/common"
	"documind.jordi.org/internal/domain/workspace"
)

var (
	ErrWorkspaceNotFound    = errors.New("workspace not found")
	ErrWorkspaceExists      = errors.New("workspace with this slug already exists")
	ErrInvalidWorkspaceName = errors.New("invalid workspace name")
	ErrInvalidSlug          = errors.New("invalid slug")
)

type Service struct {
	repo workspace.Repository
}

func NewService(repo workspace.Repository) *Service {
	return &Service{
		repo: repo,
	}
}

type CreateWorkspaceCommand struct {
	Name string
	Slug string
}

type WorkspaceDTO struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Slug      string                 `json:"slug"`
	Settings  map[string]interface{} `json:"settings"`
	CreatedAt string                 `json:"created_at"`
	UpdatedAt string                 `json:"updated_at"`
}

func (s *Service) CreateWorkspace(ctx context.Context, cmd CreateWorkspaceCommand) (*WorkspaceDTO, error) {
	if cmd.Name == "" {
		return nil, ErrInvalidWorkspaceName
	}

	if cmd.Slug == "" {
		cmd.Slug = generateSlug(cmd.Name)
	}

	slug, err := common.NewSlug(cmd.Slug)
	if err != nil {
		return nil, ErrInvalidSlug
	}

	existing, err := s.repo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrWorkspaceExists
	}

	ws := workspace.NewWorkspace(cmd.Name, slug)

	if err := s.repo.Create(ctx, ws); err != nil {
		return nil, err
	}

	return toDTO(ws), nil
}

type UpdateWorkspaceCommand struct {
	ID       string
	Name     string
	Settings map[string]interface{}
}

func (s *Service) UpdateWorkspace(ctx context.Context, cmd UpdateWorkspaceCommand) (*WorkspaceDTO, error) {
	id := common.ID(cmd.ID)

	ws, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if ws == nil {
		return nil, ErrWorkspaceNotFound
	}

	if cmd.Name != "" {
		ws.UpdateName(cmd.Name)
	}

	if cmd.Settings != nil {
		ws.UpdateSettings(cmd.Settings)
	}

	if err := s.repo.Update(ctx, ws); err != nil {
		return nil, err
	}

	return toDTO(ws), nil
}

func (s *Service) GetWorkspace(ctx context.Context, id string) (*WorkspaceDTO, error) {
	ws, err := s.repo.GetByID(ctx, common.ID(id))
	if err != nil {
		return nil, err
	}
	if ws == nil {
		return nil, ErrWorkspaceNotFound
	}

	return toDTO(ws), nil
}

func (s *Service) GetWorkspaceBySlug(ctx context.Context, slug string) (*WorkspaceDTO, error) {
	slugVO, err := common.NewSlug(slug)
	if err != nil {
		return nil, ErrInvalidSlug
	}

	ws, err := s.repo.GetBySlug(ctx, slugVO)
	if err != nil {
		return nil, err
	}
	if ws == nil {
		return nil, ErrWorkspaceNotFound
	}

	return toDTO(ws), nil
}

func (s *Service) ListWorkspaces(ctx context.Context, limit, offset int) ([]*WorkspaceDTO, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	workspaces, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	dtos := make([]*WorkspaceDTO, len(workspaces))
	for i, ws := range workspaces {
		dtos[i] = toDTO(ws)
	}

	return dtos, nil
}

func (s *Service) DeleteWorkspace(ctx context.Context, id string) error {
	ws, err := s.repo.GetByID(ctx, common.ID(id))
	if err != nil {
		return err
	}
	if ws == nil {
		return ErrWorkspaceNotFound
	}

	return s.repo.Delete(ctx, ws.ID)
}

func toDTO(ws *workspace.Workspace) *WorkspaceDTO {
	return &WorkspaceDTO{
		ID:        ws.ID.String(),
		Name:      ws.Name,
		Slug:      ws.Slug.String(),
		Settings:  ws.Settings,
		CreatedAt: ws.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: ws.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func generateSlug(name string) string {
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		return -1
	}, slug)
	return slug
}
