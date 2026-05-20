package service

import (
	"context"
	"errors"
	"sync"
	"time"

	"government-subsidy-system/backend/domain"
	"government-subsidy-system/backend/repository"
)

var ErrMissingClaim = errors.New("claimId and nationalId are required")

type DOPAClient interface {
	Verify(ctx context.Context, nationalID string) (domain.DOPAResult, error)
}

type SSOClient interface {
	Status(ctx context.Context, nationalID string) (domain.SSOResult, error)
}

type KTBClient interface {
	FinancialCheck(ctx context.Context, nationalID string) (domain.KTBResult, error)
}

type OrchestratorService struct {
	repo        repository.ClaimRepository
	projectRepo repository.ProjectRepository
	dopa        DOPAClient
	sso         SSOClient
	ktb         KTBClient
}

func NewOrchestratorService(
	repo repository.ClaimRepository,
	projectRepo repository.ProjectRepository,
	dopa DOPAClient,
	sso SSOClient,
	ktb KTBClient,
) *OrchestratorService {
	return &OrchestratorService{
		repo:        repo,
		projectRepo: projectRepo,
		dopa:        dopa,
		sso:         sso,
		ktb:         ktb,
	}
}

func (s *OrchestratorService) Orchestrate(ctx context.Context, req domain.OrchestrateRequest) (domain.DecisionResult, error) {
	if req.ClaimID == "" || req.NationalID == "" {
		return domain.DecisionResult{}, ErrMissingClaim
	}

	s.publish(req.ClaimID, domain.StatusProcessing, "started external verification")

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var criteria domain.ProjectCriteria
	hasProject := false
	if req.ProjectID != "" && s.projectRepo != nil {
		if proj, ok := s.projectRepo.Get(ctx, req.ProjectID); ok {
			criteria = proj.Criteria
			hasProject = true
		}
	}

	// Default fallback for backward compatibility
	if !hasProject {
		criteria = domain.ProjectCriteria{
			MinAge:           18,
			MaxMonthlyIncome: 30000,
			RequirePromptPay: true,
			RequireSSO:       true,
			RequireKTB:       true,
		}
	}

	var (
		wg      sync.WaitGroup
		sources domain.EligibilitySources
		errs    []string
		mu      sync.Mutex
	)

	// DOPA is always checked
	wg.Add(1)
	go func() {
		defer wg.Done()
		result, err := s.dopa.Verify(ctx, req.NationalID)
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			errs = append(errs, "DOPA verification unavailable")
			return
		}
		sources.DOPA = result
	}()

	// Query SSO only if required
	if criteria.RequireSSO {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, err := s.sso.Status(ctx, req.NationalID)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errs = append(errs, "SSO status unavailable")
				return
			}
			sources.SSO = result
		}()
	}

	// Query KTB only if required
	if criteria.RequireKTB {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, err := s.ktb.FinancialCheck(ctx, req.NationalID)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errs = append(errs, "KTB financial check unavailable")
				return
			}
			sources.KTB = result
		}()
	}

	wg.Wait()

	decision := EvaluateDecisionWithCriteria(req.ClaimID, sources, errs, criteria)
	s.repo.SaveResult(decision)
	s.publish(req.ClaimID, decision.Status, decisionMessage(decision))

	return decision, nil
}

func (s *OrchestratorService) Decision(ctx context.Context, req domain.DecisionRequest) (domain.DecisionResult, error) {
	if req.ClaimID == "" {
		return domain.DecisionResult{}, errors.New("claimId is required")
	}

	result, ok := s.repo.GetResult(req.ClaimID)
	if !ok {
		return domain.DecisionResult{}, errors.New("claim result not found")
	}

	return result, nil
}

func (s *OrchestratorService) Events(claimID string) []domain.StatusEvent {
	return s.repo.Events(claimID)
}

func (s *OrchestratorService) Subscribe(claimID string) (<-chan domain.StatusEvent, func()) {
	return s.repo.Subscribe(claimID)
}

func (s *OrchestratorService) publish(claimID string, status domain.ClaimStatus, message string) {
	s.repo.AppendEvent(domain.StatusEvent{
		ClaimID: claimID,
		Status:  status,
		Message: message,
		At:      time.Now().UTC(),
	})
}

func EvaluateDecision(claimID string, sources domain.EligibilitySources, providerErrors []string) domain.DecisionResult {
	defaultCriteria := domain.ProjectCriteria{
		MinAge:           18,
		MaxMonthlyIncome: 30000,
		RequirePromptPay: true,
		RequireSSO:       true,
		RequireKTB:       true,
	}
	return EvaluateDecisionWithCriteria(claimID, sources, providerErrors, defaultCriteria)
}

func EvaluateDecisionWithCriteria(claimID string, sources domain.EligibilitySources, providerErrors []string, criteria domain.ProjectCriteria) domain.DecisionResult {
	reasons := append([]string{}, providerErrors...)
	status := domain.StatusApproved

	if len(providerErrors) > 0 {
		status = domain.StatusPending
	}

	// 1. DOPA Verification rules
	if !sources.DOPA.Valid || !sources.DOPA.Alive || !sources.DOPA.CardActive {
		status = domain.StatusRejected
		reasons = append(reasons, "citizen identity is invalid or inactive")
	}

	minAge := criteria.MinAge
	if minAge == 0 {
		minAge = 18
	}
	if sources.DOPA.Age < minAge {
		status = domain.StatusRejected
		reasons = append(reasons, "citizen age is below minimum requirement")
	}

	if criteria.MaxAge > 0 && sources.DOPA.Age > criteria.MaxAge {
		status = domain.StatusRejected
		reasons = append(reasons, "citizen age exceeds maximum requirement")
	}

	// 2. SSO rules (only if required)
	if criteria.RequireSSO {
		if len(criteria.AllowedSSOSections) > 0 {
			allowed := false
			for _, sec := range criteria.AllowedSSOSections {
				if sources.SSO.Section == sec {
					allowed = true
					break
				}
			}
			if !allowed {
				status = domain.StatusRejected
				reasons = append(reasons, "citizen is insured under a restricted SSO section")
			}
		} else {
			// Backward compatible default behavior
			if sources.SSO.Section == "33" {
				status = domain.StatusRejected
				reasons = append(reasons, "citizen is insured under SSO section 33")
			}
		}
	}

	// 3. KTB rules (only if required)
	if criteria.RequireKTB {
		if criteria.RequirePromptPay && !sources.KTB.PromptPayLinked {
			status = domain.StatusPending
			reasons = append(reasons, "PromptPay account is not linked")
		}
		if criteria.MaxMonthlyIncome > 0 && sources.KTB.AverageMonthlyIncome > criteria.MaxMonthlyIncome {
			status = domain.StatusRejected
			reasons = append(reasons, "average monthly income exceeds threshold")
		}
		if criteria.MaxDepositTotal > 0 && sources.KTB.DepositTotal > criteria.MaxDepositTotal {
			status = domain.StatusRejected
			reasons = append(reasons, "bank deposit balance exceeds threshold")
		}
	}

	if len(reasons) == 0 {
		reasons = append(reasons, "eligible under project criteria rules")
	}

	return domain.DecisionResult{
		ClaimID: claimID,
		Status:  status,
		Reasons: reasons,
		Sources: sources,
	}
}

func decisionMessage(result domain.DecisionResult) string {
	switch result.Status {
	case domain.StatusApproved:
		return "claim approved automatically"
	case domain.StatusRejected:
		return "claim rejected by eligibility rules"
	default:
		return "claim requires follow-up"
	}
}
