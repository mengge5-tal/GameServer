package service

import (
	"GameServer/internal/application/dto"
	"GameServer/internal/domain/entity"
	"GameServer/internal/domain/repository"
)

// FriendService handles friend-related business logic
type FriendService struct {
	friendRepo repository.FriendRepository
	userRepo   repository.UserRepository
	playerRepo repository.PlayerRepository
}

// NewFriendService creates a new friend service
func NewFriendService(
	friendRepo repository.FriendRepository,
	userRepo repository.UserRepository,
	playerRepo repository.PlayerRepository,
) *FriendService {
	return &FriendService{
		friendRepo: friendRepo,
		userRepo:   userRepo,
		playerRepo: playerRepo,
	}
}

// GetFriends retrieves all friends for a user
func (s *FriendService) GetFriends(userID int) ([]*dto.FriendResponse, error) {
	friends, err := s.friendRepo.GetFriendsByUserID(userID)
	if err != nil {
		return nil, err
	}

	var response []*dto.FriendResponse
	for _, friend := range friends {
		friendUserID := friend.ToUserID
		if friend.ToUserID == userID {
			friendUserID = friend.FromUserID
		}

		// Get friend's username
		friendUser, err := s.userRepo.GetByID(friendUserID)
		if err != nil {
			continue // Skip this friend if we can't get user info
		}

		// Get friend's level
		friendPlayer, err := s.playerRepo.GetByUserID(friendUserID)
		friendLevel := 1
		if err == nil && friendPlayer != nil {
			friendLevel = friendPlayer.Level
		}

		response = append(response, &dto.FriendResponse{
			ID:             friend.ID,
			FromUserID:     friend.FromUserID,
			ToUserID:       friend.ToUserID,
			Status:         friend.Status,
			CreatedAt:      friend.CreatedAt,
			UpdatedAt:      friend.UpdatedAt,
			FriendUsername: friendUser.Username,
			FriendLevel:    friendLevel,
		})
	}

	return response, nil
}

// GetFriendRequests retrieves all pending friend requests for a user
func (s *FriendService) GetFriendRequests(userID int) ([]*dto.FriendRequestResponse, error) {
	requests, err := s.friendRepo.GetFriendRequestsByUserID(userID)
	if err != nil {
		return nil, err
	}

	var response []*dto.FriendRequestResponse
	for _, request := range requests {
		// Get requester's username
		requester, err := s.userRepo.GetByID(request.FromUserID)
		if err != nil {
			continue // Skip this request if we can't get user info
		}

		response = append(response, &dto.FriendRequestResponse{
			ID:                request.ID,
			FromUserID:        request.FromUserID,
			ToUserID:          request.ToUserID,
			Message:           request.Message,
			Status:            request.Status,
			CreatedAt:         request.CreatedAt,
			UpdatedAt:         request.UpdatedAt,
			RequesterUsername: requester.Username,
		})
	}

	return response, nil
}

// SendFriendRequest sends a friend request
func (s *FriendService) SendFriendRequest(fromUserID int, req *dto.AddFriendRequest) error {
	// Validate that users exist
	fromUser, err := s.userRepo.GetByID(fromUserID)
	if err != nil {
		return err
	}
	if fromUser == nil {
		return entity.NewDomainError("sender user not found")
	}

	toUser, err := s.userRepo.GetByID(req.ToUserID)
	if err != nil {
		return err
	}
	if toUser == nil {
		return entity.NewDomainError("target user not found")
	}

	// Check if users are already friends
	areFriends, err := s.friendRepo.AreFriends(fromUserID, req.ToUserID)
	if err != nil {
		return err
	}
	if areFriends {
		return entity.NewDomainError("users are already friends")
	}

	// Check if there's already a pending request
	hasPending, err := s.friendRepo.HasPendingRequest(fromUserID, req.ToUserID)
	if err != nil {
		return err
	}
	if hasPending {
		return entity.NewDomainError("friend request already sent")
	}

	// Create friend request
	friendRequest := &entity.FriendRequest{
		FromUserID: fromUserID,
		ToUserID:   req.ToUserID,
		Message:    req.Message,
		Status:     "pending",
	}

	return s.friendRepo.CreateFriendRequest(friendRequest)
}

// AcceptFriendRequest accepts a friend request
func (s *FriendService) AcceptFriendRequest(userID int, req *dto.FriendActionRequest) error {
	// Verify that the request exists and belongs to the user
	requests, err := s.friendRepo.GetFriendRequestsByUserID(userID)
	if err != nil {
		return err
	}

	var targetRequest *entity.FriendRequest
	for _, request := range requests {
		if request.ID == req.RequestID {
			targetRequest = request
			break
		}
	}

	if targetRequest == nil {
		return entity.NewDomainError("friend request not found")
	}

	return s.friendRepo.AcceptFriendRequest(req.RequestID)
}

// RejectFriendRequest rejects a friend request
func (s *FriendService) RejectFriendRequest(userID int, req *dto.FriendActionRequest) error {
	// Verify that the request exists and belongs to the user
	requests, err := s.friendRepo.GetFriendRequestsByUserID(userID)
	if err != nil {
		return err
	}

	var targetRequest *entity.FriendRequest
	for _, request := range requests {
		if request.ID == req.RequestID {
			targetRequest = request
			break
		}
	}

	if targetRequest == nil {
		return entity.NewDomainError("friend request not found")
	}

	return s.friendRepo.RejectFriendRequest(req.RequestID)
}

// RemoveFriend removes a friendship
func (s *FriendService) RemoveFriend(userID int, req *dto.RemoveFriendRequest) error {
	// Verify that users are friends
	areFriends, err := s.friendRepo.AreFriends(userID, req.FriendUserID)
	if err != nil {
		return err
	}
	if !areFriends {
		return entity.NewDomainError("users are not friends")
	}

	return s.friendRepo.RemoveFriend(userID, req.FriendUserID)
}

// GetFriendRanking retrieves ranking for user's friends
func (s *FriendService) GetFriendRanking(userID int) ([]*dto.FriendRankResponse, error) {
	friends, err := s.friendRepo.GetFriendsByUserID(userID)
	if err != nil {
		return nil, err
	}

	var response []*dto.FriendRankResponse
	for _, friend := range friends {
		friendUserID := friend.ToUserID
		if friend.ToUserID == userID {
			friendUserID = friend.FromUserID
		}

		// Get friend's user info
		friendUser, err := s.userRepo.GetByID(friendUserID)
		if err != nil {
			continue
		}

		// Get friend's player info
		friendPlayer, err := s.playerRepo.GetByUserID(friendUserID)
		if err != nil {
			continue
		}

		response = append(response, &dto.FriendRankResponse{
			UserID:     friendUserID,
			Username:   friendUser.Username,
			Level:      friendPlayer.Level,
			Experience: friendPlayer.Experience,
		})
	}

	return response, nil
}