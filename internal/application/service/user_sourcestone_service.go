package service

import (
	"GameServer/internal/application/dto"
	"GameServer/internal/domain/entity"
	"GameServer/internal/domain/repository"
	"database/sql"
	"fmt"
)

// UserSourceStoneService handles user source stone ownership business logic
type UserSourceStoneService struct {
	userSourceStoneRepo repository.UserSourceStoneRepository
	sourceStoneRepo     repository.SourceStoneRepository
	userRepo            repository.UserRepository
}

// NewUserSourceStoneService creates a new user source stone service
func NewUserSourceStoneService(
	userSourceStoneRepo repository.UserSourceStoneRepository,
	sourceStoneRepo repository.SourceStoneRepository,
	userRepo repository.UserRepository,
) *UserSourceStoneService {
	return &UserSourceStoneService{
		userSourceStoneRepo: userSourceStoneRepo,
		sourceStoneRepo:     sourceStoneRepo,
		userRepo:            userRepo,
	}
}

// GetUserSourceStones retrieves all source stones owned by a user
func (s *UserSourceStoneService) GetUserSourceStones(userID int, withDetails bool) (interface{}, error) {
	if userID <= 0 {
		return nil, entity.NewDomainError("user ID must be positive")
	}

	// Check if user exists
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, entity.NewDomainError("user not found")
	}

	userSourceStones, err := s.userSourceStoneRepo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	if !withDetails {
		// Return simple response without source stone details
		var response []*dto.UserSourceStoneResponse
		for _, uss := range userSourceStones {
			response = append(response, &dto.UserSourceStoneResponse{
				ID:               uss.ID,
				UserID:           uss.UserID,
				SourceStoneID:    uss.SourceStoneID,
				SourceStoneCount: uss.SourceStoneCount,
			})
		}
		return response, nil
	}

	// Return detailed response with source stone information
	var response []*dto.UserSourceStoneDetailResponse
	for _, uss := range userSourceStones {
		detailResp := &dto.UserSourceStoneDetailResponse{
			ID:               uss.ID,
			UserID:           uss.UserID,
			SourceStoneID:    uss.SourceStoneID,
			SourceStoneCount: uss.SourceStoneCount,
		}

		// Get source stone details
		sourceStone, err := s.sourceStoneRepo.GetByID(uss.SourceStoneID)
		if err != nil && err != sql.ErrNoRows {
			return nil, err
		}
		if sourceStone != nil {
			detailResp.SourceStone = &dto.SourceStoneResponse{
				SourceStoneID:      sourceStone.SourceStoneID,
				SourceStoneName:    sourceStone.SourceStoneName,
				SourceStoneQuality: sourceStone.SourceStoneQuality,
				SourceStoneEffect:  sourceStone.SourceStoneEffect,
			}
		}

		response = append(response, detailResp)
	}

	return response, nil
}

// AddUserSourceStone adds a source stone to user's inventory
func (s *UserSourceStoneService) AddUserSourceStone(req *dto.AddUserSourceStoneRequest) (*dto.UserSourceStoneResponse, error) {
	if req.UserID <= 0 {
		return nil, entity.NewDomainError("user ID must be positive")
	}
	if req.SourceStoneID <= 0 {
		return nil, entity.NewDomainError("source stone ID must be positive")
	}
	if req.SourceStoneCount <= 0 {
		return nil, entity.NewDomainError("source stone count must be positive")
	}

	// Check if user exists
	user, err := s.userRepo.GetByID(req.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, entity.NewDomainError("user not found")
	}

	// Check if source stone exists
	sourceStone, err := s.sourceStoneRepo.GetByID(req.SourceStoneID)
	if err != nil {
		return nil, err
	}
	if sourceStone == nil {
		return nil, entity.NewDomainError("source stone not found")
	}

	// Check if user already owns this source stone
	existingUserSourceStone, err := s.userSourceStoneRepo.GetByUserAndSourceStone(req.UserID, req.SourceStoneID)
	if err != nil {
		return nil, err
	}

	if existingUserSourceStone != nil {
		// User already owns this source stone, update the count
		existingUserSourceStone.SourceStoneCount += req.SourceStoneCount
		if err := s.userSourceStoneRepo.Update(existingUserSourceStone); err != nil {
			return nil, err
		}
		return &dto.UserSourceStoneResponse{
			ID:               existingUserSourceStone.ID,
			UserID:           existingUserSourceStone.UserID,
			SourceStoneID:    existingUserSourceStone.SourceStoneID,
			SourceStoneCount: existingUserSourceStone.SourceStoneCount,
		}, nil
	}

	// Create new user source stone ownership
	userSourceStone := &entity.UserSourceStone{
		UserID:           req.UserID,
		SourceStoneID:    req.SourceStoneID,
		SourceStoneCount: req.SourceStoneCount,
	}

	if err := s.userSourceStoneRepo.Create(userSourceStone); err != nil {
		return nil, err
	}

	return &dto.UserSourceStoneResponse{
		ID:               userSourceStone.ID,
		UserID:           userSourceStone.UserID,
		SourceStoneID:    userSourceStone.SourceStoneID,
		SourceStoneCount: userSourceStone.SourceStoneCount,
	}, nil
}

// UpdateUserSourceStone updates user source stone count
func (s *UserSourceStoneService) UpdateUserSourceStone(userID int, req *dto.UpdateUserSourceStoneRequest) (*dto.UserSourceStoneResponse, error) {
	if userID <= 0 {
		return nil, entity.NewDomainError("user ID must be positive")
	}
	if req.SourceStoneID <= 0 {
		return nil, entity.NewDomainError("source stone ID must be positive")
	}
	if req.SourceStoneCount < 0 {
		return nil, entity.NewDomainError("source stone count cannot be negative")
	}

	// Get existing user source stone
	userSourceStone, err := s.userSourceStoneRepo.GetByUserAndSourceStone(userID, req.SourceStoneID)
	if err != nil {
		return nil, err
	}
	if userSourceStone == nil {
		return nil, entity.NewDomainError("user source stone not found")
	}

	// If count is 0, delete the record
	if req.SourceStoneCount == 0 {
		if err := s.userSourceStoneRepo.DeleteByUserAndSourceStone(userID, req.SourceStoneID); err != nil {
			return nil, err
		}
		return &dto.UserSourceStoneResponse{
			ID:               userSourceStone.ID,
			UserID:           userSourceStone.UserID,
			SourceStoneID:    userSourceStone.SourceStoneID,
			SourceStoneCount: 0,
		}, nil
	}

	// Update the count
	userSourceStone.SourceStoneCount = req.SourceStoneCount
	if err := s.userSourceStoneRepo.Update(userSourceStone); err != nil {
		return nil, err
	}

	return &dto.UserSourceStoneResponse{
		ID:               userSourceStone.ID,
		UserID:           userSourceStone.UserID,
		SourceStoneID:    userSourceStone.SourceStoneID,
		SourceStoneCount: userSourceStone.SourceStoneCount,
	}, nil
}

// RemoveUserSourceStone removes a source stone from user's inventory
func (s *UserSourceStoneService) RemoveUserSourceStone(req *dto.RemoveUserSourceStoneRequest) error {
	if req.UserID <= 0 {
		return entity.NewDomainError("user ID must be positive")
	}
	if req.SourceStoneID <= 0 {
		return entity.NewDomainError("source stone ID must be positive")
	}

	// Check if user owns this source stone
	userSourceStone, err := s.userSourceStoneRepo.GetByUserAndSourceStone(req.UserID, req.SourceStoneID)
	if err != nil {
		return err
	}
	if userSourceStone == nil {
		return entity.NewDomainError("user does not own this source stone")
	}

	return s.userSourceStoneRepo.DeleteByUserAndSourceStone(req.UserID, req.SourceStoneID)
}

// RemoveUserSourceStoneByID removes a user source stone ownership by ID
func (s *UserSourceStoneService) RemoveUserSourceStoneByID(id int) error {
	if id <= 0 {
		return entity.NewDomainError("ID must be positive")
	}

	// Check if user source stone exists
	userSourceStone, err := s.userSourceStoneRepo.GetByID(id)
	if err != nil {
		return err
	}
	if userSourceStone == nil {
		return entity.NewDomainError("user source stone ownership not found")
	}

	return s.userSourceStoneRepo.Delete(id)
}

// CheckUserSourceStone checks if a user owns a specific source stone
func (s *UserSourceStoneService) CheckUserSourceStone(req *dto.CheckUserSourceStoneRequest) (*dto.CheckUserSourceStoneResponse, error) {
	if req.UserID <= 0 {
		return nil, entity.NewDomainError("user ID must be positive")
	}
	if req.SourceStoneID <= 0 {
		return nil, entity.NewDomainError("source stone ID must be positive")
	}

	userSourceStone, err := s.userSourceStoneRepo.GetByUserAndSourceStone(req.UserID, req.SourceStoneID)
	if err != nil {
		return nil, err
	}

	response := &dto.CheckUserSourceStoneResponse{
		UserID:          req.UserID,
		SourceStoneID:   req.SourceStoneID,
		OwnsSourceStone: userSourceStone != nil,
	}

	if userSourceStone != nil {
		response.SourceStoneCount = userSourceStone.SourceStoneCount
	}

	return response, nil
}

// BatchDeleteUserSourceStone deletes multiple user source stone ownerships
func (s *UserSourceStoneService) BatchDeleteUserSourceStone(req *dto.BatchDeleteUserSourceStoneRequest, userID int) (*dto.BatchDeleteUserSourceStoneResponse, error) {
	if len(req.SourceStoneIDs) == 0 {
		return &dto.BatchDeleteUserSourceStoneResponse{
			DeletedCount: 0,
			Message:      "No source stone IDs provided",
		}, nil
	}

	// Validate that source stone IDs are positive
	var validSourceStoneIDs []int
	for _, sourcestoneID := range req.SourceStoneIDs {
		if sourcestoneID > 0 {
			validSourceStoneIDs = append(validSourceStoneIDs, sourcestoneID)
		}
	}

	if len(validSourceStoneIDs) == 0 {
		return &dto.BatchDeleteUserSourceStoneResponse{
			DeletedCount: 0,
			Message:      "No valid source stone IDs provided",
		}, nil
	}

	// Delete from database using userID and sourcestone IDs
	deletedCount, failedIDs, err := s.userSourceStoneRepo.BatchDeleteByUserAndSourceStones(userID, validSourceStoneIDs)
	if err != nil {
		return nil, err
	}

	// Prepare response
	message := fmt.Sprintf("Successfully deleted %d user source stone items", deletedCount)
	if len(failedIDs) > 0 {
		message += fmt.Sprintf(", %d items failed to delete", len(failedIDs))
	}

	return &dto.BatchDeleteUserSourceStoneResponse{
		DeletedCount: deletedCount,
		FailedIDs:    failedIDs,
		Message:      message,
	}, nil
}