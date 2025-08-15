package service

import (
	"time"
	"GameServer/internal/application/dto"
	"GameServer/internal/domain/entity"
	"GameServer/internal/domain/repository"
)

// KillCountService handles kill count-related business logic
type KillCountService struct {
	killCountRepo repository.KillCountRepository
}

// NewKillCountService creates a new kill count service
func NewKillCountService(killCountRepo repository.KillCountRepository) *KillCountService {
	return &KillCountService{
		killCountRepo: killCountRepo,
	}
}

// GetKillCount retrieves kill count for a user and date
func (s *KillCountService) GetKillCount(req *dto.GetKillCountRequest) (*dto.KillCountResponse, error) {
	userID := req.UserID
	date := req.Date
	
	// If no date provided, use today
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}
	
	killCount, err := s.killCountRepo.GetByUserIDAndDate(userID, date)
	if err != nil {
		return nil, err
	}
	
	// If no record exists, return zeros
	if killCount == nil {
		return &dto.KillCountResponse{
			UserID: userID,
			Today:  date,
			Normal: 0,
			Elite:  0,
			Boss:   0,
			Count:  0,
		}, nil
	}
	
	return &dto.KillCountResponse{
		ID:     killCount.ID,
		UserID: killCount.UserID,
		Today:  killCount.Today,
		Normal: killCount.Normal,
		Elite:  killCount.Elite,
		Boss:   killCount.Boss,
		Count:  killCount.Count,
	}, nil
}

// GetTodayKillCount retrieves today's kill count for a user
func (s *KillCountService) GetTodayKillCount(userID int) (*dto.KillCountResponse, error) {
	killCount, err := s.killCountRepo.GetTodayKillCount(userID)
	if err != nil {
		return nil, err
	}
	
	today := time.Now().Format("2006-01-02")
	
	// If no record exists, return zeros
	if killCount == nil {
		return &dto.KillCountResponse{
			UserID: userID,
			Today:  today,
			Normal: 0,
			Elite:  0,
			Boss:   0,
			Count:  0,
		}, nil
	}
	
	return &dto.KillCountResponse{
		ID:     killCount.ID,
		UserID: killCount.UserID,
		Today:  killCount.Today,
		Normal: killCount.Normal,
		Elite:  killCount.Elite,
		Boss:   killCount.Boss,
		Count:  killCount.Count,
	}, nil
}

// UpdateKillCount updates kill count for a user (sets absolute values)
func (s *KillCountService) UpdateKillCount(userID int, req *dto.UpdateKillCountRequest) (*dto.KillCountResponse, error) {
	today := time.Now().Format("2006-01-02")
	
	// Check if record exists
	existing, err := s.killCountRepo.GetByUserIDAndDate(userID, today)
	if err != nil {
		return nil, err
	}
	
	if existing == nil {
		// Create new record
		newRecord := &entity.KillCount{
			UserID: userID,
			Today:  today,
			Normal: req.Normal,
			Elite:  req.Elite,
			Boss:   req.Boss,
		}
		
		err = s.killCountRepo.Create(newRecord)
		if err != nil {
			return nil, err
		}
		
		return &dto.KillCountResponse{
			ID:     newRecord.ID,
			UserID: newRecord.UserID,
			Today:  newRecord.Today,
			Normal: newRecord.Normal,
			Elite:  newRecord.Elite,
			Boss:   newRecord.Boss,
			Count:  newRecord.Count,
		}, nil
	} else {
		// Update existing record
		existing.Normal = req.Normal
		existing.Elite = req.Elite
		existing.Boss = req.Boss
		
		err = s.killCountRepo.Update(existing)
		if err != nil {
			return nil, err
		}
		
		return &dto.KillCountResponse{
			ID:     existing.ID,
			UserID: existing.UserID,
			Today:  existing.Today,
			Normal: existing.Normal,
			Elite:  existing.Elite,
			Boss:   existing.Boss,
			Count:  existing.Count,
		}, nil
	}
}

// IncrementKillCount increments kill count for a specific monster type
func (s *KillCountService) IncrementKillCount(userID int, req *dto.IncrementKillCountRequest) (*dto.KillCountResponse, error) {
	today := time.Now().Format("2006-01-02")
	count := req.Count
	
	// Default increment is 1
	if count <= 0 {
		count = 1
	}
	
	// Validate monster type
	validTypes := map[string]bool{
		"normal": true,
		"elite":  true,
		"boss":   true,
	}
	
	if !validTypes[req.MonsterType] {
		return nil, entity.NewDomainError("invalid monster type: must be 'normal', 'elite', or 'boss'")
	}
	
	err := s.killCountRepo.IncrementKill(userID, today, req.MonsterType, count)
	if err != nil {
		return nil, err
	}
	
	// Return updated kill count
	return s.GetTodayKillCount(userID)
}

// ResetAllKillCounts resets all users' kill counts for today (used by daily reset job)
func (s *KillCountService) ResetAllKillCounts() error {
	return s.killCountRepo.ResetAllToday()
}

// BatchIncrementKillCount increments kill counts for multiple monster types in one operation
func (s *KillCountService) BatchIncrementKillCount(userID int, req *dto.BatchIncrementKillCountRequest) (*dto.KillCountResponse, error) {
	today := time.Now().Format("2006-01-02")
	
	// Validate that at least one field is provided
	if req.Normal == nil && req.Elite == nil && req.Boss == nil {
		return nil, entity.NewDomainError("at least one monster type count must be provided")
	}
	
	// Validate that all provided counts are positive
	if req.Normal != nil && *req.Normal < 0 {
		return nil, entity.NewDomainError("normal count cannot be negative")
	}
	if req.Elite != nil && *req.Elite < 0 {
		return nil, entity.NewDomainError("elite count cannot be negative")
	}
	if req.Boss != nil && *req.Boss < 0 {
		return nil, entity.NewDomainError("boss count cannot be negative")
	}
	
	// Get or create existing record
	existing, err := s.killCountRepo.GetByUserIDAndDate(userID, today)
	if err != nil {
		return nil, err
	}
	
	if existing == nil {
		// Create new record with provided increments
		newRecord := &entity.KillCount{
			UserID: userID,
			Today:  today,
			Normal: 0,
			Elite:  0,
			Boss:   0,
		}
		
		if req.Normal != nil {
			newRecord.Normal = *req.Normal
		}
		if req.Elite != nil {
			newRecord.Elite = *req.Elite
		}
		if req.Boss != nil {
			newRecord.Boss = *req.Boss
		}
		
		err = s.killCountRepo.Create(newRecord)
		if err != nil {
			return nil, err
		}
		
		return &dto.KillCountResponse{
			ID:     newRecord.ID,
			UserID: newRecord.UserID,
			Today:  newRecord.Today,
			Normal: newRecord.Normal,
			Elite:  newRecord.Elite,
			Boss:   newRecord.Boss,
			Count:  newRecord.Count,
		}, nil
	} else {
		// Update existing record with increments
		if req.Normal != nil {
			existing.Normal += *req.Normal
		}
		if req.Elite != nil {
			existing.Elite += *req.Elite
		}
		if req.Boss != nil {
			existing.Boss += *req.Boss
		}
		
		err = s.killCountRepo.Update(existing)
		if err != nil {
			return nil, err
		}
		
		return &dto.KillCountResponse{
			ID:     existing.ID,
			UserID: existing.UserID,
			Today:  existing.Today,
			Normal: existing.Normal,
			Elite:  existing.Elite,
			Boss:   existing.Boss,
			Count:  existing.Count,
		}, nil
	}
}

// DeleteKillCount deletes a kill count record by ID
func (s *KillCountService) DeleteKillCount(id int) error {
	return s.killCountRepo.Delete(id)
}

// GetKillRanking retrieves kill count ranking top N players
func (s *KillCountService) GetKillRanking(req *dto.GetKillRankingRequest) ([]*dto.KillRankingResponse, error) {
	limit := req.Limit
	if limit <= 0 {
		limit = 100 // Default to top 100
	}
	
	entries, err := s.killCountRepo.GetKillRanking(limit)
	if err != nil {
		return nil, err
	}
	
	var rankings []*dto.KillRankingResponse
	for _, entry := range entries {
		rankings = append(rankings, &dto.KillRankingResponse{
			UserID:   entry.UserID,
			Username: entry.Username,
			Level:    entry.Level,
			Count:    entry.Count,
			Rank:     entry.Rank,
		})
	}
	
	return rankings, nil
}

// GetUserKillRank retrieves specific user's kill count ranking
func (s *KillCountService) GetUserKillRank(req *dto.GetUserKillRankRequest) (*dto.UserKillRankResponse, error) {
	userID := req.UserID
	if userID <= 0 {
		return nil, entity.NewDomainError("user ID is required")
	}
	
	entry, err := s.killCountRepo.GetUserKillRank(userID)
	if err != nil {
		return nil, err
	}
	
	return &dto.UserKillRankResponse{
		UserID:   entry.UserID,
		Username: entry.Username,
		Level:    entry.Level,
		Count:    entry.Count,
		Rank:     entry.Rank,
	}, nil
}