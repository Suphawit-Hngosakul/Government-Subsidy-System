package adapter

import (
	"context"

	"government-subsidy-system/backend/domain"
)

type MockDOPAAdapter struct{}

func NewMockDOPAAdapter() MockDOPAAdapter {
	return MockDOPAAdapter{}
}

func (MockDOPAAdapter) Verify(ctx context.Context, nationalID string) (domain.DOPAResult, error) {
	return domain.DOPAResult{
		Valid:      true,
		Age:        35,
		Alive:      true,
		CardActive: true,
	}, nil
}
