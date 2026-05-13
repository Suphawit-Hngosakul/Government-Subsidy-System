package service

import (
	"context"
	"errors"
	"testing"

	"government-subsidy-system/backend/domain"
	"government-subsidy-system/backend/repository"
)

func newProjectServiceForTest() *ProjectService {
	return NewProjectService(repository.NewMemoryProjectRepository())
}

func ptrString(s string) *string { return &s }
func ptrBool(b bool) *bool       { return &b }

func TestProjectServiceCreatesProjectWithGeneratedID(t *testing.T) {
	svc := newProjectServiceForTest()

	result, err := svc.Create(context.Background(), domain.ProjectInput{Name: "Energy Subsidy"})
	if err != nil {
		t.Fatalf("create returned error: %v", err)
	}
	if result.ID == "" {
		t.Fatal("expected generated project id")
	}
	if result.Criteria.MinAge != defaultMinAge {
		t.Fatalf("expected default MinAge %d, got %d", defaultMinAge, result.Criteria.MinAge)
	}
}

func TestProjectServiceCreateAlwaysGeneratesServerSideID(t *testing.T) {
	svc := newProjectServiceForTest()

	first, _ := svc.Create(context.Background(), domain.ProjectInput{Name: "A"})
	second, _ := svc.Create(context.Background(), domain.ProjectInput{Name: "B"})

	if first.ID == second.ID {
		t.Fatalf("expected distinct generated ids, got %s twice", first.ID)
	}
}

func TestProjectServiceRejectsEmptyName(t *testing.T) {
	svc := newProjectServiceForTest()

	_, err := svc.Create(context.Background(), domain.ProjectInput{})
	if !errors.Is(err, ErrProjectNameRequired) {
		t.Fatalf("expected ErrProjectNameRequired, got %v", err)
	}
}

func TestProjectServiceListReturnsCreatedProjects(t *testing.T) {
	svc := newProjectServiceForTest()
	ctx := context.Background()

	if _, err := svc.Create(ctx, domain.ProjectInput{Name: "A"}); err != nil {
		t.Fatalf("create A failed: %v", err)
	}
	if _, err := svc.Create(ctx, domain.ProjectInput{Name: "B"}); err != nil {
		t.Fatalf("create B failed: %v", err)
	}

	list := svc.List(ctx)
	if len(list) != 2 {
		t.Fatalf("expected 2 projects, got %d", len(list))
	}
}

func TestProjectServiceUpdateMergesOnlyProvidedFields(t *testing.T) {
	svc := newProjectServiceForTest()
	ctx := context.Background()

	created, err := svc.Create(ctx, domain.ProjectInput{
		Name:        "Old",
		Description: "original",
		Active:      true,
		Criteria:    domain.ProjectCriteria{MaxMonthlyIncome: 15000, RequirePromptPay: true},
	})
	if err != nil {
		t.Fatalf("seed create failed: %v", err)
	}

	updated, err := svc.Update(ctx, created.ID, domain.ProjectUpdate{
		Name: ptrString("New"),
	})
	if err != nil {
		t.Fatalf("update returned error: %v", err)
	}
	if updated.Name != "New" {
		t.Fatalf("expected name New, got %s", updated.Name)
	}
	if updated.Description != "original" {
		t.Fatalf("expected description preserved, got %s", updated.Description)
	}
	if !updated.Active {
		t.Fatal("expected active preserved as true")
	}
	if updated.Criteria.MaxMonthlyIncome != 15000 {
		t.Fatalf("expected criteria preserved, got %v", updated.Criteria.MaxMonthlyIncome)
	}
	if !updated.CreatedAt.Equal(created.CreatedAt) {
		t.Fatal("expected CreatedAt preserved on update")
	}
}

func TestProjectServiceUpdateRejectsEmptyName(t *testing.T) {
	svc := newProjectServiceForTest()
	ctx := context.Background()
	created, _ := svc.Create(ctx, domain.ProjectInput{Name: "Old"})

	_, err := svc.Update(ctx, created.ID, domain.ProjectUpdate{Name: ptrString("")})
	if !errors.Is(err, ErrProjectNameRequired) {
		t.Fatalf("expected ErrProjectNameRequired, got %v", err)
	}
}

func TestProjectServiceUpdateUnknownReturnsNotFound(t *testing.T) {
	svc := newProjectServiceForTest()

	_, err := svc.Update(context.Background(), "missing", domain.ProjectUpdate{Name: ptrString("X")})
	if !errors.Is(err, repository.ErrProjectNotFound) {
		t.Fatalf("expected ErrProjectNotFound, got %v", err)
	}
}

func TestProjectServiceUpdateCanToggleActiveFalse(t *testing.T) {
	svc := newProjectServiceForTest()
	ctx := context.Background()
	created, _ := svc.Create(ctx, domain.ProjectInput{Name: "X", Active: true})

	updated, err := svc.Update(ctx, created.ID, domain.ProjectUpdate{Active: ptrBool(false)})
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if updated.Active {
		t.Fatal("expected active to be toggled to false")
	}
}

func TestProjectServiceDeleteRemovesProject(t *testing.T) {
	svc := newProjectServiceForTest()
	ctx := context.Background()
	created, err := svc.Create(ctx, domain.ProjectInput{Name: "Temp"})
	if err != nil {
		t.Fatalf("seed create failed: %v", err)
	}

	if err := svc.Delete(ctx, created.ID); err != nil {
		t.Fatalf("delete returned error: %v", err)
	}
	if _, err := svc.Get(ctx, created.ID); !errors.Is(err, repository.ErrProjectNotFound) {
		t.Fatalf("expected ErrProjectNotFound after delete, got %v", err)
	}
}

func TestProjectServiceDeleteUnknownReturnsNotFound(t *testing.T) {
	svc := newProjectServiceForTest()

	err := svc.Delete(context.Background(), "missing")
	if !errors.Is(err, repository.ErrProjectNotFound) {
		t.Fatalf("expected ErrProjectNotFound, got %v", err)
	}
}

func TestProjectServiceEmptyIDPathsReturnsErrProjectIDRequired(t *testing.T) {
	svc := newProjectServiceForTest()
	ctx := context.Background()

	if _, err := svc.Get(ctx, ""); !errors.Is(err, ErrProjectIDRequired) {
		t.Fatalf("Get(\"\"): expected ErrProjectIDRequired, got %v", err)
	}
	if _, err := svc.Update(ctx, "", domain.ProjectUpdate{Name: ptrString("X")}); !errors.Is(err, ErrProjectIDRequired) {
		t.Fatalf("Update(\"\"): expected ErrProjectIDRequired, got %v", err)
	}
	if err := svc.Delete(ctx, ""); !errors.Is(err, ErrProjectIDRequired) {
		t.Fatalf("Delete(\"\"): expected ErrProjectIDRequired, got %v", err)
	}
}
