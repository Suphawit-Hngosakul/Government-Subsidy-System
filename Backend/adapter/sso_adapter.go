package adapter

import (
	"context"

	"government-subsidy-system/backend/domain"
)

type MockSSOAdapter struct{}

func NewMockSSOAdapter() MockSSOAdapter {
	return MockSSOAdapter{}
}

func (MockSSOAdapter) Status(ctx context.Context, nationalID string) (domain.SSOResult, error) {
	return domain.SSOResult{
		Section:            "40",
		ContributionMonths: 12,
	}, nil
}
