package adapter

import (
	"context"

	"government-subsidy-system/backend/domain"
)

type MockKTBAdapter struct{}

func NewMockKTBAdapter() MockKTBAdapter {
	return MockKTBAdapter{}
}

func (MockKTBAdapter) FinancialCheck(ctx context.Context, nationalID string) (domain.KTBResult, error) {
	return domain.KTBResult{
		DepositTotal:         12000,
		AverageMonthlyIncome: 15000,
		PromptPayLinked:      true,
	}, nil
}
