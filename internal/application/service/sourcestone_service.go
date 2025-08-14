package service

import (
	"GameServer/internal/application/dto"
	"GameServer/internal/domain/entity"
	"GameServer/internal/domain/repository"
)

// SourceStoneService handles source stone business logic
type SourceStoneService struct {
	sourceStoneRepo repository.SourceStoneRepository
}

// NewSourceStoneService creates a new source stone service
func NewSourceStoneService(sourceStoneRepo repository.SourceStoneRepository) *SourceStoneService {
	return &SourceStoneService{
		sourceStoneRepo: sourceStoneRepo,
	}
}

// GetSourceStoneByID retrieves a source stone by ID
func (s *SourceStoneService) GetSourceStoneByID(sourcestoneID int) (*dto.SourceStoneResponse, error) {
	if sourcestoneID <= 0 {
		return nil, entity.NewDomainError("source stone ID must be positive")
	}

	sourceStone, err := s.sourceStoneRepo.GetByID(sourcestoneID)
	if err != nil {
		return nil, err
	}
	if sourceStone == nil {
		return nil, entity.NewDomainError("source stone not found")
	}

	return &dto.SourceStoneResponse{
		SourceStoneID:      sourceStone.SourceStoneID,
		SourceStoneName:    sourceStone.SourceStoneName,
		SourceStoneQuality: sourceStone.SourceStoneQuality,
		SourceStoneEffect:  sourceStone.SourceStoneEffect,
	}, nil
}

// GetAllSourceStones retrieves all source stones
func (s *SourceStoneService) GetAllSourceStones() ([]*dto.SourceStoneResponse, error) {
	sourceStones, err := s.sourceStoneRepo.GetAll()
	if err != nil {
		return nil, err
	}

	var response []*dto.SourceStoneResponse
	for _, stone := range sourceStones {
		response = append(response, &dto.SourceStoneResponse{
			SourceStoneID:      stone.SourceStoneID,
			SourceStoneName:    stone.SourceStoneName,
			SourceStoneQuality: stone.SourceStoneQuality,
			SourceStoneEffect:  stone.SourceStoneEffect,
		})
	}

	return response, nil
}