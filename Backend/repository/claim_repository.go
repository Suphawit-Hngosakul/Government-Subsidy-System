package repository

import (
	"sync"

	"government-subsidy-system/backend/domain"
)

type ClaimRepository interface {
	SaveResult(result domain.DecisionResult)
	GetResult(claimID string) (domain.DecisionResult, bool)
	AppendEvent(event domain.StatusEvent)
	Events(claimID string) []domain.StatusEvent
	Subscribe(claimID string) (<-chan domain.StatusEvent, func())
}

type MemoryClaimRepository struct {
	mu          sync.RWMutex
	results     map[string]domain.DecisionResult
	events      map[string][]domain.StatusEvent
	subscribers map[string]map[chan domain.StatusEvent]struct{}
}

func NewMemoryClaimRepository() *MemoryClaimRepository {
	return &MemoryClaimRepository{
		results:     make(map[string]domain.DecisionResult),
		events:      make(map[string][]domain.StatusEvent),
		subscribers: make(map[string]map[chan domain.StatusEvent]struct{}),
	}
}

func (r *MemoryClaimRepository) SaveResult(result domain.DecisionResult) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.results[result.ClaimID] = result
}

func (r *MemoryClaimRepository) GetResult(claimID string) (domain.DecisionResult, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result, ok := r.results[claimID]
	return result, ok
}

func (r *MemoryClaimRepository) AppendEvent(event domain.StatusEvent) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.events[event.ClaimID] = append(r.events[event.ClaimID], event)

	for subscriber := range r.subscribers[event.ClaimID] {
		select {
		case subscriber <- event:
		default:
		}
	}
}

func (r *MemoryClaimRepository) Events(claimID string) []domain.StatusEvent {
	r.mu.RLock()
	defer r.mu.RUnlock()

	events := r.events[claimID]
	copied := make([]domain.StatusEvent, len(events))
	copy(copied, events)
	return copied
}

func (r *MemoryClaimRepository) Subscribe(claimID string) (<-chan domain.StatusEvent, func()) {
	ch := make(chan domain.StatusEvent, 8)

	r.mu.Lock()
	if r.subscribers[claimID] == nil {
		r.subscribers[claimID] = make(map[chan domain.StatusEvent]struct{})
	}
	r.subscribers[claimID][ch] = struct{}{}
	r.mu.Unlock()

	unsubscribe := func() {
		r.mu.Lock()
		defer r.mu.Unlock()

		delete(r.subscribers[claimID], ch)
		if len(r.subscribers[claimID]) == 0 {
			delete(r.subscribers, claimID)
		}
		close(ch)
	}

	return ch, unsubscribe
}
