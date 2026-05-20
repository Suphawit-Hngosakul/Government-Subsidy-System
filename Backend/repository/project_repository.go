package repository

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"

	"government-subsidy-system/backend/domain"
)

var (
	ErrProjectNotFound      = errors.New("project not found")
	ErrProjectAlreadyExists = errors.New("project id already exists")
)

type ProjectRepository interface {
	Create(ctx context.Context, project domain.Project) (domain.Project, error)
	Get(ctx context.Context, id string) (domain.Project, bool)
	List(ctx context.Context) []domain.Project
	Update(ctx context.Context, project domain.Project) (domain.Project, error)
	Delete(ctx context.Context, id string) error
}

type MemoryProjectRepository struct {
	mu       sync.RWMutex
	projects map[string]domain.Project
}

func NewMemoryProjectRepository() *MemoryProjectRepository {
	return &MemoryProjectRepository{projects: make(map[string]domain.Project)}
}

func (r *MemoryProjectRepository) Create(_ context.Context, project domain.Project) (domain.Project, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.projects[project.ID]; exists {
		return domain.Project{}, ErrProjectAlreadyExists
	}

	now := time.Now().UTC()
	project.CreatedAt = now
	project.UpdatedAt = now
	r.projects[project.ID] = project
	return project, nil
}

func (r *MemoryProjectRepository) Get(_ context.Context, id string) (domain.Project, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	project, ok := r.projects[id]
	return project, ok
}

func (r *MemoryProjectRepository) List(_ context.Context) []domain.Project {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]domain.Project, 0, len(r.projects))
	for _, p := range r.projects {
		list = append(list, p)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].CreatedAt.Before(list[j].CreatedAt)
	})
	return list
}

func (r *MemoryProjectRepository) Update(_ context.Context, project domain.Project) (domain.Project, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, ok := r.projects[project.ID]
	if !ok {
		return domain.Project{}, ErrProjectNotFound
	}

	project.CreatedAt = existing.CreatedAt
	project.UpdatedAt = time.Now().UTC()
	r.projects[project.ID] = project
	return project, nil
}

func (r *MemoryProjectRepository) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.projects[id]; !ok {
		return ErrProjectNotFound
	}
	delete(r.projects, id)
	return nil
}
