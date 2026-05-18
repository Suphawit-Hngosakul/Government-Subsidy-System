package repository

import (
	"fmt"
	"sync"
	"time"

	"government-subsidy-system/backend/domain"
)

type CitizenRepository interface {
	FindByNationalID(nationalID string) (*domain.Citizen, error)
	Create(citizen *domain.Citizen) error
	UpdateKYCStatus(nationalID string, status domain.KYCStatus) error
	UpdateLaserCode(nationalID string, laserCode string) error
}

type memoryCitizenRepository struct {
	mu       sync.RWMutex
	citizens map[string]*domain.Citizen
	counter  int
}

func NewMemoryCitizenRepository() CitizenRepository {
	return &memoryCitizenRepository{
		citizens: make(map[string]*domain.Citizen),
	}
}

func (r *memoryCitizenRepository) FindByNationalID(nationalID string) (*domain.Citizen, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.citizens[nationalID]
	if !ok {
		return nil, fmt.Errorf("citizen not found")
	}
	return c, nil
}

func (r *memoryCitizenRepository) Create(citizen *domain.Citizen) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.citizens[citizen.NationalID]; exists {
		return fmt.Errorf("citizen already registered")
	}
	r.counter++
	citizen.ID = fmt.Sprintf("citizen-%03d", r.counter)
	citizen.CreatedAt = time.Now()
	r.citizens[citizen.NationalID] = citizen
	return nil
}

func (r *memoryCitizenRepository) UpdateKYCStatus(nationalID string, status domain.KYCStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.citizens[nationalID]
	if !ok {
		return fmt.Errorf("citizen not found")
	}
	c.KYCStatus = status
	return nil
}

func (r *memoryCitizenRepository) UpdateLaserCode(nationalID string, laserCode string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.citizens[nationalID]
	if !ok {
		return fmt.Errorf("citizen not found")
	}
	c.LaserCode = laserCode
	return nil
}
