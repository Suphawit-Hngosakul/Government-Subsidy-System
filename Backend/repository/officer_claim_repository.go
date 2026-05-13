package repository

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"

	"government-subsidy-system/backend/domain"
)

var ErrOfficerClaimNotFound = errors.New("claim not found")

type OfficerClaimRepository interface {
	Get(ctx context.Context, claimID string) (domain.OfficerClaim, bool)
	ListByStatus(ctx context.Context, status domain.OfficerClaimStatus) []domain.OfficerClaim
	Update(ctx context.Context, claim domain.OfficerClaim) (domain.OfficerClaim, error)
}

type MemoryOfficerClaimRepository struct {
	mu     sync.RWMutex
	claims map[string]domain.OfficerClaim
}

func NewMemoryOfficerClaimRepository() *MemoryOfficerClaimRepository {
	r := &MemoryOfficerClaimRepository{claims: make(map[string]domain.OfficerClaim)}
	r.seedMockClaims()
	return r
}

func (r *MemoryOfficerClaimRepository) seedMockClaims() {
	now := time.Now().UTC()
	seeds := []domain.OfficerClaim{
		{
			ClaimID:     "claim-001",
			NationalID:  "1101700203451",
			ProjectID:   "proj-energy",
			Status:      domain.OfficerStatusPending,
			SubmittedAt: now.Add(-2 * time.Hour),
		},
		{
			ClaimID:     "claim-002",
			NationalID:  "1101700203452",
			ProjectID:   "proj-energy",
			Status:      domain.OfficerStatusPending,
			SubmittedAt: now.Add(-1 * time.Hour),
		},
		{
			ClaimID:     "claim-003",
			NationalID:  "1101700203453",
			ProjectID:   "proj-water",
			Status:      domain.OfficerStatusPending,
			SubmittedAt: now.Add(-30 * time.Minute),
		},
	}
	for _, c := range seeds {
		r.claims[c.ClaimID] = c
	}
}

func (r *MemoryOfficerClaimRepository) Get(_ context.Context, claimID string) (domain.OfficerClaim, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	claim, ok := r.claims[claimID]
	return claim, ok
}

func (r *MemoryOfficerClaimRepository) ListByStatus(_ context.Context, status domain.OfficerClaimStatus) []domain.OfficerClaim {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]domain.OfficerClaim, 0)
	for _, c := range r.claims {
		if c.Status == status {
			list = append(list, c)
		}
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].SubmittedAt.Before(list[j].SubmittedAt)
	})
	return list
}

func (r *MemoryOfficerClaimRepository) Update(_ context.Context, claim domain.OfficerClaim) (domain.OfficerClaim, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.claims[claim.ClaimID]; !ok {
		return domain.OfficerClaim{}, ErrOfficerClaimNotFound
	}
	r.claims[claim.ClaimID] = claim
	return claim, nil
}
