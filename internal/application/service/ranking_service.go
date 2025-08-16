package service

import (
	"GameServer/internal/application/dto"
	"GameServer/internal/domain/entity"
	"GameServer/internal/domain/repository"
)

// RankingService handles player ranking business logic
type RankingService struct {
	playerRepo repository.PlayerRepository
}

// NewRankingService creates a new ranking service
func NewRankingService(playerRepo repository.PlayerRepository) *RankingService {
	return &RankingService{
		playerRepo: playerRepo,
	}
}

// GetPlayerRanking retrieves player ranking by type with limit
func (s *RankingService) GetPlayerRanking(req *dto.PlayerRankingRequest) (*dto.GetPlayerRankingResponse, error) {
	// Validate rank type
	validTypes := []string{"level", "experience", "gamelevel", "bloodenergy"}
	isValid := false
	for _, validType := range validTypes {
		if req.RankType == validType {
			isValid = true
			break
		}
	}
	if !isValid {
		return nil, entity.NewDomainError("invalid rank type")
	}

	// Set default limit if not specified or invalid
	limit := req.Limit
	if limit <= 0 || limit > 100 {
		limit = 50 // Default to top 50
	}

	// Get rankings from repository
	rankings, err := s.playerRepo.GetPlayerRanking(req.RankType, limit)
	if err != nil {
		return nil, err
	}

	// Convert to response DTOs
	var responseRankings []*dto.PlayerRankingResponse
	for _, ranking := range rankings {
		responseRankings = append(responseRankings, &dto.PlayerRankingResponse{
			UserID:   ranking.UserID,
			Username: ranking.Username,
			Value:    ranking.Value,
			Position: ranking.Position,
		})
	}

	return &dto.GetPlayerRankingResponse{
		RankType: req.RankType,
		Rankings: responseRankings,
	}, nil
}

// GetUserRank retrieves a specific user's rank for a given rank type
func (s *RankingService) GetUserRank(userID int, rankType string) (*dto.UserRankResponse, error) {
	// Validate rank type
	validTypes := []string{"level", "experience", "gamelevel", "bloodenergy"}
	isValid := false
	for _, validType := range validTypes {
		if rankType == validType {
			isValid = true
			break
		}
	}
	if !isValid {
		return nil, entity.NewDomainError("invalid rank type")
	}

	// Get user ranking
	ranking, err := s.playerRepo.GetUserRank(userID, rankType)
	if err != nil {
		return nil, err
	}
	if ranking == nil {
		return nil, entity.NewDomainError("user ranking not found")
	}

	return &dto.UserRankResponse{
		UserID:   ranking.UserID,
		Username: ranking.Username,
		RankType: rankType,
		Value:    ranking.Value,
		Position: ranking.Position,
	}, nil
}