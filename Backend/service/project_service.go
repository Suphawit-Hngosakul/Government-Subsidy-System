package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"

	"government-subsidy-system/backend/domain"
	"government-subsidy-system/backend/repository"
)

const defaultMinAge = 18

var (
	ErrProjectNameRequired = errors.New("project name is required")
	ErrProjectIDRequired   = errors.New("project id is required")
)

type ProjectService struct {
	repo repository.ProjectRepository
}

func NewProjectService(repo repository.ProjectRepository) *ProjectService {
	return &ProjectService{repo: repo}
}

func (s *ProjectService) Create(ctx context.Context, input domain.ProjectInput) (domain.Project, error) {
	if input.Name == "" {
		return domain.Project{}, ErrProjectNameRequired
	}

	project := domain.Project{
		ID:          newProjectID(),
		Name:        input.Name,
		Description: input.Description,
		Active:      input.Active,
		Criteria:    input.Criteria,
	}
	if project.Criteria.MinAge == 0 {
		project.Criteria.MinAge = defaultMinAge
	}
	return s.repo.Create(ctx, project)
}

func (s *ProjectService) List(ctx context.Context) []domain.Project {
	return s.repo.List(ctx)
}

func (s *ProjectService) Get(ctx context.Context, id string) (domain.Project, error) {
	if id == "" {
		return domain.Project{}, ErrProjectIDRequired
	}
	project, ok := s.repo.Get(ctx, id)
	if !ok {
		return domain.Project{}, repository.ErrProjectNotFound
	}
	return project, nil
}

func (s *ProjectService) Update(ctx context.Context, id string, patch domain.ProjectUpdate) (domain.Project, error) {
	if id == "" {
		return domain.Project{}, ErrProjectIDRequired
	}

	existing, ok := s.repo.Get(ctx, id)
	if !ok {
		return domain.Project{}, repository.ErrProjectNotFound
	}

	if patch.Name != nil {
		if *patch.Name == "" {
			return domain.Project{}, ErrProjectNameRequired
		}
		existing.Name = *patch.Name
	}
	if patch.Description != nil {
		existing.Description = *patch.Description
	}
	if patch.Active != nil {
		existing.Active = *patch.Active
	}
	if patch.Criteria != nil {
		existing.Criteria = *patch.Criteria
	}

	return s.repo.Update(ctx, existing)
}

func (s *ProjectService) Delete(ctx context.Context, id string) error {
	if id == "" {
		return ErrProjectIDRequired
	}
	return s.repo.Delete(ctx, id)
}

func newProjectID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return "proj-" + hex.EncodeToString(b)
}
