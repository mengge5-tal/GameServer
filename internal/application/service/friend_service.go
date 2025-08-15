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
			FriendID:       friendUserID,
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

		// Get requester's level
		requesterPlayer, err := s.playerRepo.GetByUserID(request.FromUserID)
		requesterLevel := 1
		if err == nil && requesterPlayer != nil {
			requesterLevel = requesterPlayer.Level
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
			RequesterLevel:    requesterLevel,
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

// AcceptFriendRequestByUserID accepts a friend request by user IDs
func (s *FriendService) AcceptFriendRequestByUserID(toUserID, fromUserID int) error {
	// Find the friend request by user IDs
	requests, err := s.friendRepo.GetFriendRequestsByUserID(toUserID)
	if err != nil {
		return err
	}

	var requestID int
	found := false
	for _, req := range requests {
		if req.FromUserID == fromUserID && req.ToUserID == toUserID && req.Status == "pending" {
			requestID = req.ID
			found = true
			break
		}
	}

	if !found {
		return entity.NewDomainError("friend request not found or already processed")
	}

	// AcceptFriendRequest will now delete the request after processing
	return s.friendRepo.AcceptFriendRequest(requestID)
}

// RejectFriendRequestByUserID rejects a friend request by user IDs
func (s *FriendService) RejectFriendRequestByUserID(toUserID, fromUserID int) error {
	// Find the friend request by user IDs
	requests, err := s.friendRepo.GetFriendRequestsByUserID(toUserID)
	if err != nil {
		return err
	}

	var requestID int
	found := false
	for _, req := range requests {
		if req.FromUserID == fromUserID && req.ToUserID == toUserID && req.Status == "pending" {
			requestID = req.ID
			found = true
			break
		}
	}

	if !found {
		return entity.NewDomainError("friend request not found or already processed")
	}

	// RejectFriendRequest will now delete the request after processing
	return s.friendRepo.RejectFriendRequest(requestID)
}

// GetRecommendedFriends retrieves recommended friends for a user
func (s *FriendService) GetRecommendedFriends(userID int) ([]*dto.RecommendedFriendResponse, error) {
	// Get user's current level
	currentUser, err := s.playerRepo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}
	
	userLevel := 1
	if currentUser != nil {
		userLevel = currentUser.Level
	}
	
	const maxRecommendations = 6
	var recommendations []*entity.UserWithLevel
	
	// Strategy 1: Level difference within 5, online users
	if len(recommendations) < maxRecommendations {
		candidates, err := s.friendRepo.GetRecommendedFriends(userID, userLevel, 5, maxRecommendations, true)
		if err == nil {
			recommendations = append(recommendations, candidates...)
		}
	}
	
	// Strategy 2: Level difference within 10, online users
	if len(recommendations) < maxRecommendations {
		needed := maxRecommendations - len(recommendations)
		candidates, err := s.friendRepo.GetRecommendedFriends(userID, userLevel, 10, needed, true)
		if err == nil {
			// Filter out already recommended users
			for _, candidate := range candidates {
				alreadyRecommended := false
				for _, existing := range recommendations {
					if existing.UserID == candidate.UserID {
						alreadyRecommended = true
						break
					}
				}
				if !alreadyRecommended {
					recommendations = append(recommendations, candidate)
					if len(recommendations) >= maxRecommendations {
						break
					}
				}
			}
		}
	}
	
	// Strategy 3: All online users (no level restriction)
	if len(recommendations) < maxRecommendations {
		needed := maxRecommendations - len(recommendations)
		candidates, err := s.friendRepo.GetRecommendedFriends(userID, userLevel, 0, needed, true)
		if err == nil {
			// Filter out already recommended users
			for _, candidate := range candidates {
				alreadyRecommended := false
				for _, existing := range recommendations {
					if existing.UserID == candidate.UserID {
						alreadyRecommended = true
						break
					}
				}
				if !alreadyRecommended {
					recommendations = append(recommendations, candidate)
					if len(recommendations) >= maxRecommendations {
						break
					}
				}
			}
		}
	}
	
	// Strategy 4: All users (no restrictions)
	if len(recommendations) < maxRecommendations {
		needed := maxRecommendations - len(recommendations)
		candidates, err := s.friendRepo.GetRecommendedFriends(userID, userLevel, 0, needed, false)
		if err == nil {
			// Filter out already recommended users
			for _, candidate := range candidates {
				alreadyRecommended := false
				for _, existing := range recommendations {
					if existing.UserID == candidate.UserID {
						alreadyRecommended = true
						break
					}
				}
				if !alreadyRecommended {
					recommendations = append(recommendations, candidate)
					if len(recommendations) >= maxRecommendations {
						break
					}
				}
			}
		}
	}
	
	// Convert to response DTOs
	var response []*dto.RecommendedFriendResponse
	for _, user := range recommendations {
		response = append(response, &dto.RecommendedFriendResponse{
			UserID:   user.UserID,
			Username: user.Username,
			Level:    user.Level,
		})
	}
	
	return response, nil
}