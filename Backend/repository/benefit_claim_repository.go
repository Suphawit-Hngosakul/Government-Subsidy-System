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

	// Bridge to sync data with Officer feature
	officerRepo OfficerClaimRepository
}

func NewMemoryBenefitClaimRepository(officerRepo OfficerClaimRepository) *MemoryBenefitClaimRepository {
	return &MemoryBenefitClaimRepository{
		claims:      make(map[string]domain.Claim),
		officerRepo: officerRepo,
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

	// --- Bridge: Sync to Officer Repository ---
	if r.officerRepo != nil {
		officerClaim := domain.OfficerClaim{
			ClaimID:     claim.ID,
			NationalID:  claim.NationalID,
			ProjectID:   claim.ProjectID,
			Status:      domain.OfficerStatusPending, // All new claims wait for Orchestrator/Officer
			SubmittedAt: claim.SubmittedAt,
		}
		// Since OfficerRepo.Update acts as an insert/update in memory, we might need a direct insert.
		// For the existing mock repo, Update checks existence. Let's add it directly if we must,
		// or rely on a new method. Wait, OfficerRepo currently returns ErrOfficerClaimNotFound on Update if not exists.
		// Let's modify the Update behavior in Officer Repo or add a Create method to it.
		// Alternatively, we can cast it to the concrete type and insert directly.
		if memoryRepo, ok := r.officerRepo.(*MemoryOfficerClaimRepository); ok {
			memoryRepo.mu.Lock()
			memoryRepo.claims[claim.ID] = officerClaim
			memoryRepo.mu.Unlock()
		}
	}
	// ------------------------------------------

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
