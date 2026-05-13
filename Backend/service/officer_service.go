package service

import (
	"context"
	"errors"
	"log"
	"time"

	"government-subsidy-system/backend/domain"
	"government-subsidy-system/backend/repository"
)

var (
	ErrOfficerIDRequired    = errors.New("officerId is required")
	ErrRejectReasonRequired = errors.New("rejection reason is required")
	ErrClaimAlreadyDecided  = errors.New("claim already decided")
	ErrClaimIDRequired      = errors.New("claimId is required")
)

type OrchestratorClient interface {
	GetDecision(ctx context.Context, claimID string) (domain.EligibilityResult, error)
}

type OfficerService struct {
	repo         repository.OfficerClaimRepository
	orchestrator OrchestratorClient
}

func NewOfficerService(repo repository.OfficerClaimRepository, orchestrator OrchestratorClient) *OfficerService {
	return &OfficerService{repo: repo, orchestrator: orchestrator}
}

func (s *OfficerService) ListPending(ctx context.Context) []domain.OfficerClaim {
	return s.repo.ListByStatus(ctx, domain.OfficerStatusPending)
}

func (s *OfficerService) Detail(ctx context.Context, claimID string) (domain.OfficerClaimDetail, error) {
	if claimID == "" {
		return domain.OfficerClaimDetail{}, ErrClaimIDRequired
	}

	claim, ok := s.repo.Get(ctx, claimID)
	if !ok {
		return domain.OfficerClaimDetail{}, repository.ErrOfficerClaimNotFound
	}

	detail := domain.OfficerClaimDetail{Claim: claim}

	result, err := s.orchestrator.GetDecision(ctx, claimID)
	if err != nil {
		log.Printf("officer: orchestrator unavailable for claim %s: %v", claimID, err)
		return detail, nil
	}
	if result.ClaimID == "" {
		return detail, nil
	}

	detail.Eligibility = &result
	return detail, nil
}

func (s *OfficerService) Approve(ctx context.Context, claimID string, input domain.OfficerDecisionInput) (domain.OfficerClaim, error) {
	return s.decide(ctx, claimID, input, domain.OfficerStatusApproved, false)
}

func (s *OfficerService) Reject(ctx context.Context, claimID string, input domain.OfficerDecisionInput) (domain.OfficerClaim, error) {
	return s.decide(ctx, claimID, input, domain.OfficerStatusRejected, true)
}

func (s *OfficerService) decide(ctx context.Context, claimID string, input domain.OfficerDecisionInput, status domain.OfficerClaimStatus, reasonRequired bool) (domain.OfficerClaim, error) {
	if claimID == "" {
		return domain.OfficerClaim{}, ErrClaimIDRequired
	}
	if input.OfficerID == "" {
		return domain.OfficerClaim{}, ErrOfficerIDRequired
	}
	if reasonRequired && input.Reason == "" {
		return domain.OfficerClaim{}, ErrRejectReasonRequired
	}

	claim, ok := s.repo.Get(ctx, claimID)
	if !ok {
		return domain.OfficerClaim{}, repository.ErrOfficerClaimNotFound
	}
	if claim.Status != domain.OfficerStatusPending {
		return domain.OfficerClaim{}, ErrClaimAlreadyDecided
	}

	claim.Status = status
	claim.OfficerDecision = &domain.OfficerDecision{
		OfficerID: input.OfficerID,
		Reason:    input.Reason,
		DecidedAt: time.Now().UTC(),
	}

	return s.repo.Update(ctx, claim)
}
