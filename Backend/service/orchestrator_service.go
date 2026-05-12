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
	repo repository.ClaimRepository
	dopa DOPAClient
	sso  SSOClient
	ktb  KTBClient
}

func NewOrchestratorService(repo repository.ClaimRepository, dopa DOPAClient, sso SSOClient, ktb KTBClient) *OrchestratorService {
	return &OrchestratorService{repo: repo, dopa: dopa, sso: sso, ktb: ktb}
}

func (s *OrchestratorService) Orchestrate(ctx context.Context, req domain.OrchestrateRequest) (domain.DecisionResult, error) {
	if req.ClaimID == "" || req.NationalID == "" {
		return domain.DecisionResult{}, ErrMissingClaim
	}

	s.publish(req.ClaimID, domain.StatusProcessing, "started external verification")

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var (
		wg      sync.WaitGroup
		sources domain.EligibilitySources
		errs    []string
		mu      sync.Mutex
	)

	wg.Add(3)
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

	wg.Wait()

	decision := EvaluateDecision(req.ClaimID, sources, errs)
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
	reasons := append([]string{}, providerErrors...)
	status := domain.StatusApproved

	if len(providerErrors) > 0 {
		status = domain.StatusPending
	}
	if !sources.DOPA.Valid || !sources.DOPA.Alive || !sources.DOPA.CardActive {
		status = domain.StatusRejected
		reasons = append(reasons, "citizen identity is invalid or inactive")
	}
	if sources.DOPA.Age < 18 {
		status = domain.StatusRejected
		reasons = append(reasons, "citizen age is below minimum requirement")
	}
	if sources.SSO.Section == "33" {
		status = domain.StatusRejected
		reasons = append(reasons, "citizen is insured under SSO section 33")
	}
	if !sources.KTB.PromptPayLinked {
		status = domain.StatusPending
		reasons = append(reasons, "PromptPay account is not linked")
	}
	if sources.KTB.AverageMonthlyIncome > 30000 {
		status = domain.StatusRejected
		reasons = append(reasons, "average monthly income exceeds threshold")
	}
	if len(reasons) == 0 {
		reasons = append(reasons, "eligible by mock subsidy rules")
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

type MockDOPAClient struct{}

func (MockDOPAClient) Verify(ctx context.Context, nationalID string) (domain.DOPAResult, error) {
	return domain.DOPAResult{Valid: true, Age: 35, Alive: true, CardActive: true}, nil
}

type MockSSOClient struct{}

func (MockSSOClient) Status(ctx context.Context, nationalID string) (domain.SSOResult, error) {
	return domain.SSOResult{Section: "40", ContributionMonths: 12}, nil
}

type MockKTBClient struct{}

func (MockKTBClient) FinancialCheck(ctx context.Context, nationalID string) (domain.KTBResult, error) {
	return domain.KTBResult{DepositTotal: 12000, AverageMonthlyIncome: 15000, PromptPayLinked: true}, nil
}
