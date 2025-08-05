package service

import (
	"GameServer/internal/application/dto"
	"GameServer/internal/domain/entity"
	"GameServer/internal/domain/repository"
)

// RankingService handles ranking-related business logic
type RankingService struct {
	rankingRepo repository.RankingRepository
	userRepo    repository.UserRepository
	playerRepo  repository.PlayerRepository
}

// NewRankingService creates a new ranking service
func NewRankingService(
	rankingRepo repository.RankingRepository,
	userRepo repository.UserRepository,
	playerRepo repository.PlayerRepository,
) *RankingService {
	return &RankingService{
		rankingRepo: rankingRepo,
		userRepo:    userRepo,
		playerRepo:  playerRepo,
	}
}

// GetRanking retrieves ranking by type
func (s *RankingService) GetRanking(req *dto.GetRankingRequest) ([]*dto.RankingResponse, error) {
	// Set default limit if not specified
	limit := req.Limit
	if limit <= 0 || limit > 100 {
		limit = 50 // Default to top 50
	}

	// Validate rank type
	validTypes := []string{"level", "experience", "equipment_power"}
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

	// Get rankings from repository
	rankings, err := s.rankingRepo.GetRankingByType(req.RankType, limit)
	if err != nil {
		return nil, err
	}

	// Convert to response DTOs and populate usernames
	var response []*dto.RankingResponse
	for _, ranking := range rankings {
		// Get username
		user, err := s.userRepo.GetByID(ranking.UserID)
		if err != nil {
			continue // Skip if user not found
		}

		response = append(response, &dto.RankingResponse{
			ID:           ranking.ID,
			UserID:       ranking.UserID,
			Username:     user.Username,
			RankType:     ranking.RankType,
			RankValue:    ranking.RankValue,
			RankPosition: ranking.RankPosition,
			UpdatedAt:    ranking.UpdatedAt,
		})
	}

	return response, nil
}

// GetUserRanking retrieves specific user's ranking
func (s *RankingService) GetUserRanking(userID int, rankType string) (*dto.UserRankingResponse, error) {
	// Validate rank type
	validTypes := []string{"level", "experience", "equipment_power"}
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
	ranking, err := s.rankingRepo.GetUserRanking(userID, rankType)
	if err != nil {
		return nil, err
	}
	if ranking == nil {
		return nil, entity.NewDomainError("user ranking not found")
	}

	// Get username
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	return &dto.UserRankingResponse{
		UserID:       ranking.UserID,
		Username:     user.Username,
		RankType:     ranking.RankType,
		RankValue:    ranking.RankValue,
		RankPosition: ranking.RankPosition,
		UpdatedAt:    ranking.UpdatedAt,
	}, nil
}

// UpdateUserRankings updates all rankings for a user based on current player info
func (s *RankingService) UpdateUserRankings(userID int) error {
	// Get player info
	playerInfo, err := s.playerRepo.GetByUserID(userID)
	if err != nil {
		return err
	}
	if playerInfo == nil {
		return entity.NewDomainError("player info not found")
	}

	// Update level ranking
	if err := s.rankingRepo.UpdateUserRanking(userID, "level", playerInfo.Level); err != nil {
		return err
	}

	// Update experience ranking
	if err := s.rankingRepo.UpdateUserRanking(userID, "experience", playerInfo.Experience); err != nil {
		return err
	}

	// TODO: Calculate equipment power and update equipment_power ranking
	// For now, we'll use a placeholder value
	equipmentPower := s.calculateEquipmentPower(userID)
	if err := s.rankingRepo.UpdateUserRanking(userID, "equipment_power", equipmentPower); err != nil {
		return err
	}

	return nil
}

// RefreshAllRankings recalculates all ranking positions
func (s *RankingService) RefreshAllRankings() error {
	rankTypes := []string{"level", "experience", "equipment_power"}
	
	for _, rankType := range rankTypes {
		if err := s.rankingRepo.RefreshRankings(rankType); err != nil {
			return err
		}
	}
	
	return nil
}

// calculateEquipmentPower calculates total equipment power for a user
func (s *RankingService) calculateEquipmentPower(userID int) int {
	// This is a placeholder implementation
	// In a real game, you would calculate based on user's equipment
	// For now, return a default value
	return 100
}