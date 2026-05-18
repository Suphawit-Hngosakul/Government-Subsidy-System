package repository

import "sync"

type TokenRepository interface {
	Save(token string, citizenID string)
	FindCitizenID(token string) (string, bool)
	Revoke(token string)
}

type memoryTokenRepository struct {
	mu     sync.RWMutex
	tokens map[string]string // token -> citizenID
}

func NewMemoryTokenRepository() TokenRepository {
	return &memoryTokenRepository{
		tokens: make(map[string]string),
	}
}

func (r *memoryTokenRepository) Save(token, citizenID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tokens[token] = citizenID
}

func (r *memoryTokenRepository) FindCitizenID(token string) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id, ok := r.tokens[token]
	return id, ok
}

func (r *memoryTokenRepository) Revoke(token string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.tokens, token)
}
