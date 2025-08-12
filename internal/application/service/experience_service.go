package service

import (
	"GameServer/internal/domain/entity"
	"GameServer/internal/domain/repository"
)

// ExperienceService handles experience-related business logic
type ExperienceService struct {
	experienceRepo repository.ExperienceRepository
}

// NewExperienceService creates a new experience service
func NewExperienceService(experienceRepo repository.ExperienceRepository) *ExperienceService {
	return &ExperienceService{
		experienceRepo: experienceRepo,
	}
}

// GetByLevel retrieves experience info by level
func (s *ExperienceService) GetByLevel(level int) (*entity.Experience, error) {
	if level <= 0 {
		return nil, entity.NewDomainError("level must be positive")
	}

	return s.experienceRepo.GetByLevel(level)
}

// GetAllLevels retrieves all experience levels
func (s *ExperienceService) GetAllLevels() ([]*entity.Experience, error) {
	return s.experienceRepo.GetAllLevels()
}