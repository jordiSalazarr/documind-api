package project

import (
	"context"
	"errors"

	"documind.jordi.org/internal/domain/common"
	"documind.jordi.org/internal/domain/knowledge"
)

type Service struct {
	projectRepo knowledge.ProjectRepository
	areaRepo    knowledge.ServiceRepository
}

func NewService(projectRepo knowledge.ProjectRepository, areaRepo knowledge.ServiceRepository) *Service {
	return &Service{
		projectRepo: projectRepo,
		areaRepo:    areaRepo,
	}
}

// Project operations

type CreateProjectInput struct {
	WorkspaceID common.ID
	Name        string
	Slug        common.Slug
	Description string
	CreatedBy   common.ID
}

func (s *Service) CreateProject(ctx context.Context, input CreateProjectInput) (*knowledge.Project, error) {
	if input.Name == "" {
		return nil, errors.New("name is required")
	}

	project := knowledge.NewProject(
		input.WorkspaceID,
		input.Name,
		input.Slug,
		input.Description,
		input.CreatedBy,
	)

	if err := s.projectRepo.Create(ctx, project); err != nil {
		return nil, err
	}

	return project, nil
}

func (s *Service) GetProject(ctx context.Context, id common.ID) (*knowledge.Project, error) {
	return s.projectRepo.GetByID(ctx, id)
}

func (s *Service) GetProjectBySlug(ctx context.Context, workspaceID common.ID, slug common.Slug) (*knowledge.Project, error) {
	return s.projectRepo.GetBySlug(ctx, workspaceID, slug)
}

func (s *Service) ListProjects(ctx context.Context, workspaceID common.ID) ([]*knowledge.Project, error) {
	return s.projectRepo.ListByWorkspace(ctx, workspaceID)
}

type UpdateProjectInput struct {
	ID          common.ID
	Name        string
	Description string
	UpdatedBy   common.ID
}

func (s *Service) UpdateProject(ctx context.Context, input UpdateProjectInput) (*knowledge.Project, error) {
	project, err := s.projectRepo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	project.Update(input.Name, input.Description, input.UpdatedBy)

	if err := s.projectRepo.Update(ctx, project); err != nil {
		return nil, err
	}

	return project, nil
}

func (s *Service) DeleteProject(ctx context.Context, id common.ID) error {
	return s.projectRepo.Delete(ctx, id)
}

// Area operations

type CreateAreaInput struct {
	WorkspaceID common.ID
	ProjectID   common.ID
	Name        string
	Slug        common.Slug
	Description string
	CreatedBy   common.ID
}

func (s *Service) CreateArea(ctx context.Context, input CreateAreaInput) (*knowledge.Service, error) {
	if input.Name == "" {
		return nil, errors.New("name is required")
	}

	exists, err := s.projectRepo.Exists(ctx, input.ProjectID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("project not found")
	}

	service := knowledge.NewService(
		input.WorkspaceID,
		input.ProjectID,
		input.Name,
		input.Slug,
		input.Description,
		input.CreatedBy,
	)

	if err := s.areaRepo.Create(ctx, service); err != nil {
		return nil, err
	}

	return service, nil
}

func (s *Service) GetArea(ctx context.Context, id common.ID) (*knowledge.Service, error) {
	return s.areaRepo.GetByID(ctx, id)
}

func (s *Service) GetAreaBySlug(ctx context.Context, projectID common.ID, slug common.Slug) (*knowledge.Service, error) {
	return s.areaRepo.GetBySlug(ctx, projectID, slug)
}

func (s *Service) ListAreas(ctx context.Context, projectID common.ID) ([]*knowledge.Service, error) {
	return s.areaRepo.ListByProject(ctx, projectID)
}

type UpdateAreaInput struct {
	ID          common.ID
	Name        string
	Description string
	UpdatedBy   common.ID
}

func (s *Service) UpdateArea(ctx context.Context, input UpdateAreaInput) (*knowledge.Service, error) {
	service, err := s.areaRepo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	service.Update(input.Name, input.Description, input.UpdatedBy)

	if err := s.areaRepo.Update(ctx, service); err != nil {
		return nil, err
	}

	return service, nil
}

func (s *Service) DeleteArea(ctx context.Context, id common.ID) error {
	return s.areaRepo.Delete(ctx, id)
}
