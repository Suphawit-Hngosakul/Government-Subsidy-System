package repository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"

	"government-subsidy-system/backend/domain"
)

var ErrClaimNotFound = errors.New("claim not found")

type BenefitClaimRepository interface {
	Create(ctx context.Context, claim domain.Claim) (domain.Claim, error)
	GetByID(ctx context.Context, claimID string) (domain.Claim, error)
	GetByCitizen(ctx context.Context, nationalID string) ([]domain.Claim, error)
	Update(ctx context.Context, claim domain.Claim) (domain.Claim, error)
	UpdateStatus(ctx context.Context, claimID string, status domain.ClaimStatus) error
	UpdateEligibility(ctx context.Context, claimID string, eligibility domain.EligibilityResult) error
}

type MemoryBenefitClaimRepository struct {
	mu     sync.RWMutex
	claims map[string]domain.Claim
}

func NewMemoryBenefitClaimRepository() *MemoryBenefitClaimRepository {
	return &MemoryBenefitClaimRepository{
		claims: make(map[string]domain.Claim),
	}
}

func (r *MemoryBenefitClaimRepository) Create(ctx context.Context, claim domain.Claim) (domain.Claim, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	claim.ID = newClaimID()
	claim.SubmittedAt = time.Now().UTC()
	claim.UpdatedAt = time.Now().UTC()
	claim.Status = domain.StatusProcessing

	r.claims[claim.ID] = claim
	return claim, nil
}

func (r *MemoryBenefitClaimRepository) GetByID(ctx context.Context, claimID string) (domain.Claim, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	claim, ok := r.claims[claimID]
	if !ok {
		return domain.Claim{}, ErrClaimNotFound
	}
	return claim, nil
}

func (r *MemoryBenefitClaimRepository) GetByCitizen(ctx context.Context, nationalID string) ([]domain.Claim, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var claims []domain.Claim
	for _, claim := range r.claims {
		if claim.NationalID == nationalID {
			claims = append(claims, claim)
		}
	}
	return claims, nil
}

func (r *MemoryBenefitClaimRepository) Update(ctx context.Context, claim domain.Claim) (domain.Claim, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.claims[claim.ID]; !ok {
		return domain.Claim{}, ErrClaimNotFound
	}

	claim.UpdatedAt = time.Now().UTC()
	r.claims[claim.ID] = claim
	return claim, nil
}

func (r *MemoryBenefitClaimRepository) UpdateStatus(ctx context.Context, claimID string, status domain.ClaimStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	claim, ok := r.claims[claimID]
	if !ok {
		return ErrClaimNotFound
	}

	claim.Status = status
	claim.UpdatedAt = time.Now().UTC()
	r.claims[claimID] = claim
	return nil
}

func (r *MemoryBenefitClaimRepository) UpdateEligibility(ctx context.Context, claimID string, eligibility domain.EligibilityResult) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	claim, ok := r.claims[claimID]
	if !ok {
		return ErrClaimNotFound
	}

	claim.Eligibility = &eligibility
	claim.Status = eligibility.Status
	claim.UpdatedAt = time.Now().UTC()
	r.claims[claimID] = claim
	return nil
}

func newClaimID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return "claim-" + hex.EncodeToString(b)
}
