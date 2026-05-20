package repository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"

	"government-subsidy-system/backend/domain"
)

type AuditRepository interface {
	Append(ctx context.Context, entry domain.AuditEntry) (domain.AuditEntry, error)
	List(ctx context.Context, limit int) []domain.AuditEntry
}

type MemoryAuditRepository struct {
	mu      sync.RWMutex
	entries []domain.AuditEntry
}

func NewMemoryAuditRepository() *MemoryAuditRepository {
	return &MemoryAuditRepository{}
}

func (r *MemoryAuditRepository) Append(_ context.Context, entry domain.AuditEntry) (domain.AuditEntry, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if entry.ID == "" {
		entry.ID = newAuditID()
	}
	if entry.At.IsZero() {
		entry.At = time.Now().UTC()
	}
	r.entries = append(r.entries, entry)
	return entry, nil
}

func (r *MemoryAuditRepository) List(_ context.Context, limit int) []domain.AuditEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	total := len(r.entries)
	if limit <= 0 || limit > total {
		limit = total
	}

	result := make([]domain.AuditEntry, 0, limit)
	for i := total - 1; i >= 0 && len(result) < limit; i-- {
		result = append(result, r.entries[i])
	}
	return result
}

func newAuditID() string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	return "audit-" + hex.EncodeToString(b)
}
