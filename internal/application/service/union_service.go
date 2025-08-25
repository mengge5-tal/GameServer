package service

import (
	"GameServer/internal/application/dto"
	"GameServer/internal/domain/entity"
	"GameServer/internal/domain/repository"
	"GameServer/internal/infrastructure/cache"
	"fmt"
	"time"
)

// UnionService handles union business logic
type UnionService struct {
	unionRepo         repository.UnionRepository
	memberRepo        repository.UnionMemberRepository
	requestRepo       repository.UnionRequestRepository
	experienceRepo    repository.UnionExperienceRepository
	playerRepo        repository.PlayerRepository
	userRepo          repository.UserRepository
	cacheService      cache.CacheService
}

// NewUnionService creates a new union service
func NewUnionService(
	unionRepo repository.UnionRepository,
	memberRepo repository.UnionMemberRepository,
	requestRepo repository.UnionRequestRepository,
	experienceRepo repository.UnionExperienceRepository,
	playerRepo repository.PlayerRepository,
	userRepo repository.UserRepository,
	cacheService cache.CacheService,
) *UnionService {
	return &UnionService{
		unionRepo:      unionRepo,
		memberRepo:     memberRepo,
		requestRepo:    requestRepo,
		experienceRepo: experienceRepo,
		playerRepo:     playerRepo,
		userRepo:       userRepo,
		cacheService:   cacheService,
	}
}

// GetMyUnionInfo retrieves user's current union information
func (s *UnionService) GetMyUnionInfo(userID int) (*dto.UnionResponse, error) {
	// Check if user is in any union
	member, err := s.memberRepo.GetByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("获取用户工会成员信息失败: %w", err)
	}
	
	if member == nil {
		return nil, nil // User is not in any union
	}
	
	// Get union details
	union, err := s.unionRepo.GetByID(member.UnionID)
	if err != nil {
		return nil, fmt.Errorf("获取工会信息失败: %w", err)
	}
	
	if union == nil {
		// Clean up orphaned member record
		s.memberRepo.DeleteByUserID(userID)
		return nil, nil
	}
	
	return &dto.UnionResponse{
		UnionID:           union.UnionID,
		UnionName:         union.UnionName,
		ChairpersonID:     union.ChairpersonID,
		ChairpersonName:   union.ChairpersonName,
		ChairpersonLevel:  union.ChairpersonLevel,
		UnionLevel:        union.UnionLevel,
		UnionMembers:      union.UnionMembers,
		Experience:        union.Experience,
		CreatedTime:       union.CreatedTime,
		UnionDesc:         union.UnionDesc,
	}, nil
}

// CreateUnion creates a new union
func (s *UnionService) CreateUnion(req *dto.CreateUnionRequest) (*dto.UnionResponse, error) {
	// Validate input
	if len(req.UnionName) < 2 || len(req.UnionName) > 100 {
		return nil, fmt.Errorf("工会名称长度必须在2-100字符之间")
	}
	
	// Check if user is already in a union
	isInUnion, err := s.memberRepo.IsUserInUnion(req.ChairpersonID)
	if err != nil {
		return nil, fmt.Errorf("检查用户工会状态失败: %w", err)
	}
	
	if isInUnion {
		return nil, fmt.Errorf("您已经加入了工会，请先退出当前工会")
	}
	
	// Check if union name already exists
	exists, err := s.unionRepo.Exists(req.UnionName)
	if err != nil {
		return nil, fmt.Errorf("检查工会名称失败: %w", err)
	}
	
	if exists {
		return nil, fmt.Errorf("工会名称已存在，请选择其他名称")
	}
	
	// Get user info to verify blood energy and get user details
	user, err := s.userRepo.GetByID(req.ChairpersonID)
	if err != nil {
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}
	
	if user == nil {
		return nil, fmt.Errorf("用户不存在")
	}
	
	// Get player info to check blood energy and level
	playerInfo, err := s.playerRepo.GetByUserID(req.ChairpersonID)
	if err != nil {
		return nil, fmt.Errorf("获取玩家信息失败: %w", err)
	}
	
	if playerInfo == nil {
		return nil, fmt.Errorf("玩家信息不存在")
	}
	
	// Check if user has enough blood energy (1000 required)
	if playerInfo.BloodEnergy < 1000 {
		return nil, fmt.Errorf("创建工会需要1000点血能量，您当前只有%d点", playerInfo.BloodEnergy)
	}
	
	// Start transaction-like operations
	// Deduct blood energy first
	newBloodEnergy := playerInfo.BloodEnergy - 1000
	err = s.playerRepo.UpdateBloodEnergy(req.ChairpersonID, newBloodEnergy)
	if err != nil {
		return nil, fmt.Errorf("扣除血能量失败: %w", err)
	}
	
	// Create union entity
	union := &entity.Union{
		UnionName:        req.UnionName,
		ChairpersonID:    req.ChairpersonID,
		ChairpersonName:  req.ChairpersonName,
		ChairpersonLevel: playerInfo.Level,
		UnionLevel:       1,
		UnionMembers:     1,
		Experience:       0,
		CreatedTime:      time.Now(),
		UnionDesc:        req.UnionDesc,
	}
	
	// Validate union entity
	if err := union.Validate(); err != nil {
		// Rollback blood energy
		s.playerRepo.UpdateBloodEnergy(req.ChairpersonID, playerInfo.BloodEnergy)
		return nil, fmt.Errorf("工会信息验证失败: %w", err)
	}
	
	// Create union
	err = s.unionRepo.Create(union)
	if err != nil {
		// Rollback blood energy
		s.playerRepo.UpdateBloodEnergy(req.ChairpersonID, playerInfo.BloodEnergy)
		return nil, fmt.Errorf("创建工会失败: %w", err)
	}
	
	// Add creator as chairperson member
	member := &entity.UnionMember{
		UnionID:     union.UnionID,
		UnionName:   union.UnionName,
		MemberID:    req.ChairpersonID,
		MemberLevel: playerInfo.Level,
		JoinedTime:  time.Now(),
		RoleID:      entity.UnionRoleLeader, // 2 = Leader/Chairperson
	}
	
	err = s.memberRepo.Create(member)
	if err != nil {
		// Rollback: delete union and restore blood energy
		s.unionRepo.Delete(union.UnionID)
		s.playerRepo.UpdateBloodEnergy(req.ChairpersonID, playerInfo.BloodEnergy)
		return nil, fmt.Errorf("添加工会创建者失败: %w", err)
	}
	
	// Clear user's union cache if exists
	s.cacheService.Delete(fmt.Sprintf("user_union:%d", req.ChairpersonID))
	
	// Return created union information
	return &dto.UnionResponse{
		UnionID:           union.UnionID,
		UnionName:         union.UnionName,
		ChairpersonID:     union.ChairpersonID,
		ChairpersonName:   union.ChairpersonName,
		ChairpersonLevel:  union.ChairpersonLevel,
		UnionLevel:        union.UnionLevel,
		UnionMembers:      union.UnionMembers,
		Experience:        union.Experience,
		CreatedTime:       union.CreatedTime,
		UnionDesc:         union.UnionDesc,
	}, nil
}

// JoinUnion handles user joining a union (creates a request)
func (s *UnionService) JoinUnion(req *dto.JoinUnionRequest) error {
	// Check if user is already in a union
	isInUnion, err := s.memberRepo.IsUserInUnion(req.ApplicantID)
	if err != nil {
		return fmt.Errorf("检查用户工会状态失败: %w", err)
	}
	
	if isInUnion {
		return fmt.Errorf("您已经加入了工会")
	}
	
	// Check if union exists
	union, err := s.unionRepo.GetByID(req.UnionID)
	if err != nil {
		return fmt.Errorf("获取工会信息失败: %w", err)
	}
	
	if union == nil {
		return fmt.Errorf("工会不存在")
	}
	
	// Check if user already has a pending request to this union
	hasPending, err := s.requestRepo.HasPendingRequest(req.ApplicantID, req.UnionID)
	if err != nil {
		return fmt.Errorf("检查待处理申请失败: %w", err)
	}
	
	if hasPending {
		return fmt.Errorf("您已经向该工会提交了申请，请等待处理")
	}
	
	// Get applicant info
	user, err := s.userRepo.GetByID(req.ApplicantID)
	if err != nil {
		return fmt.Errorf("获取申请人信息失败: %w", err)
	}
	
	playerInfo, err := s.playerRepo.GetByUserID(req.ApplicantID)
	if err != nil {
		return fmt.Errorf("获取申请人玩家信息失败: %w", err)
	}
	
	// Create union request
	request := &entity.UnionRequest{
		UnionID:        req.UnionID,
		ApplicantID:    req.ApplicantID,
		ApplicantName:  user.Username,
		ApplicantLevel: playerInfo.Level,
		ChairpersonID:  union.ChairpersonID,
		RequestStatus:  entity.UnionRequestStatusPending,
		RequestTime:    time.Now(),
	}
	
	err = s.requestRepo.Create(request)
	if err != nil {
		return fmt.Errorf("创建加入申请失败: %w", err)
	}
	
	return nil
}

// GetRecommendedUnions retrieves recommended unions
func (s *UnionService) GetRecommendedUnions(limit int) (*dto.UnionListResponse, error) {
	if limit <= 0 || limit > 20 {
		limit = 6 // Default to 6 as specified
	}
	
	unions, err := s.unionRepo.GetRecommended(limit)
	if err != nil {
		return nil, fmt.Errorf("获取推荐工会失败: %w", err)
	}
	
	summaries := make([]dto.UnionSummary, 0, len(unions))
	for _, union := range unions {
		summaries = append(summaries, dto.UnionSummary{
			UnionID:           union.UnionID,
			UnionName:         union.UnionName,
			ChairpersonName:   union.ChairpersonName,
			ChairpersonLevel:  union.ChairpersonLevel,
			UnionLevel:        union.UnionLevel,
			UnionMembers:      union.UnionMembers,
			UnionDesc:         union.UnionDesc,
		})
	}
	
	return &dto.UnionListResponse{
		Unions: summaries,
		Total:  len(summaries),
	}, nil
}

// ProcessUnionRequest processes a union join request (approve/reject)
func (s *UnionService) ProcessUnionRequest(req *dto.ProcessUnionRequestDTO) error {
	// Get the request details
	request, err := s.requestRepo.GetByID(req.RequestID)
	if err != nil {
		return fmt.Errorf("获取申请信息失败: %w", err)
	}
	
	if request == nil {
		return fmt.Errorf("申请不存在")
	}
	
	// Verify that the user is the chairperson of the union
	union, err := s.unionRepo.GetByID(request.UnionID)
	if err != nil {
		return fmt.Errorf("获取工会信息失败: %w", err)
	}
	
	if union == nil {
		return fmt.Errorf("工会不存在")
	}
	
	if union.ChairpersonID != req.ChairpersonID {
		return fmt.Errorf("只有工会会长可以处理申请")
	}
	
	// Check if request is still pending
	if request.RequestStatus != entity.UnionRequestStatusPending {
		return fmt.Errorf("该申请已被处理")
	}
	
	// If approving, check if applicant is still not in a union
	if req.Status == entity.UnionRequestStatusApproved {
		isInUnion, err := s.memberRepo.IsUserInUnion(request.ApplicantID)
		if err != nil {
			return fmt.Errorf("检查申请人工会状态失败: %w", err)
		}
		
		if isInUnion {
			// Update request status to rejected since user is already in a union
			s.requestRepo.ProcessRequest(req.RequestID, entity.UnionRequestStatusRejected)
			return fmt.Errorf("申请人已加入其他工会")
		}
		
		// Add member to union
		member := &entity.UnionMember{
			UnionID:     request.UnionID,
			UnionName:   union.UnionName,
			MemberID:    request.ApplicantID,
			MemberLevel: request.ApplicantLevel,
			JoinedTime:  time.Now(),
			RoleID:      entity.UnionRoleMember,
		}
		
		err = s.memberRepo.Create(member)
		if err != nil {
			return fmt.Errorf("添加工会成员失败: %w", err)
		}
		
		// Update union member count
		err = s.unionRepo.IncrementMemberCount(request.UnionID)
		if err != nil {
			// Rollback member creation
			s.memberRepo.DeleteByUserID(request.ApplicantID)
			return fmt.Errorf("更新工会成员数量失败: %w", err)
		}
		
		// Clear cache
		s.cacheService.Delete(fmt.Sprintf("user_union:%d", request.ApplicantID))
	}
	
	// Update request status
	err = s.requestRepo.ProcessRequest(req.RequestID, req.Status)
	if err != nil {
		return fmt.Errorf("更新申请状态失败: %w", err)
	}
	
	return nil
}

// GetUnionInfo gets detailed union information including members
func (s *UnionService) GetUnionInfo(unionID int) (*dto.UnionResponse, error) {
	union, err := s.unionRepo.GetByID(unionID)
	if err != nil {
		return nil, fmt.Errorf("获取工会信息失败: %w", err)
	}
	
	if union == nil {
		return nil, nil
	}
	
	return &dto.UnionResponse{
		UnionID:           union.UnionID,
		UnionName:         union.UnionName,
		ChairpersonID:     union.ChairpersonID,
		ChairpersonName:   union.ChairpersonName,
		ChairpersonLevel:  union.ChairpersonLevel,
		UnionLevel:        union.UnionLevel,
		UnionMembers:      union.UnionMembers,
		Experience:        union.Experience,
		CreatedTime:       union.CreatedTime,
		UnionDesc:         union.UnionDesc,
	}, nil
}

// LeaveUnion handles user leaving their current union
func (s *UnionService) LeaveUnion(userID int) error {
	// Get user's current union membership
	member, err := s.memberRepo.GetByUserID(userID)
	if err != nil {
		return fmt.Errorf("获取用户工会信息失败: %w", err)
	}
	
	if member == nil {
		return fmt.Errorf("您不在任何工会中")
	}
	
	// Check if user is the chairperson
	if member.RoleID == entity.UnionRoleLeader {
		// Get union info to check member count
		union, err := s.unionRepo.GetByID(member.UnionID)
		if err != nil {
			return fmt.Errorf("获取工会信息失败: %w", err)
		}
		
		if union.UnionMembers > 1 {
			return fmt.Errorf("工会会长不能直接退出工会，请先转让会长职位或解散工会")
		}
		
		// If chairperson is the only member, dissolve the union
		return s.DismissUnion(&dto.DismissUnionRequest{
			UnionID:       member.UnionID,
			ChairpersonID: userID,
		})
	}
	
	// Remove member from union
	err = s.memberRepo.DeleteByUserID(userID)
	if err != nil {
		return fmt.Errorf("退出工会失败: %w", err)
	}
	
	// Update union member count
	err = s.unionRepo.DecrementMemberCount(member.UnionID)
	if err != nil {
		// This is not critical, log but don't fail
		fmt.Printf("Warning: Failed to update union member count: %v", err)
	}
	
	// Clear cache
	s.cacheService.Delete(fmt.Sprintf("user_union:%d", userID))
	
	return nil
}

// DismissUnion dissolves/dismisses a union
func (s *UnionService) DismissUnion(req *dto.DismissUnionRequest) error {
	// Get union info and verify chairperson
	union, err := s.unionRepo.GetByID(req.UnionID)
	if err != nil {
		return fmt.Errorf("获取工会信息失败: %w", err)
	}
	
	if union == nil {
		return fmt.Errorf("工会不存在")
	}
	
	if union.ChairpersonID != req.ChairpersonID {
		return fmt.Errorf("只有工会会长可以解散工会")
	}
	
	// Get all members to clear their cache
	members, err := s.memberRepo.GetByUnionID(req.UnionID)
	if err != nil {
		return fmt.Errorf("获取工会成员失败: %w", err)
	}
	
	// Delete all union members first
	for _, member := range members {
		err = s.memberRepo.Delete(member.ID)
		if err != nil {
			fmt.Printf("Warning: Failed to delete member %d: %v", member.MemberID, err)
		}
		// Clear member's cache
		s.cacheService.Delete(fmt.Sprintf("user_union:%d", member.MemberID))
	}
	
	// Delete all pending requests for this union
	requests, err := s.requestRepo.GetByUnionID(req.UnionID)
	if err == nil {
		for _, request := range requests {
			s.requestRepo.Delete(request.ID)
		}
	}
	
	// Delete the union
	err = s.unionRepo.Delete(req.UnionID)
	if err != nil {
		return fmt.Errorf("解散工会失败: %w", err)
	}
	
	return nil
}

// GetUnionRanking gets union ranking
func (s *UnionService) GetUnionRanking(limit int) (*dto.UnionListResponse, error) {
	if limit <= 0 || limit > 50 {
		limit = 10 // Default limit
	}
	
	unions, err := s.unionRepo.GetRanking(limit)
	if err != nil {
		return nil, fmt.Errorf("获取工会排行榜失败: %w", err)
	}
	
	summaries := make([]dto.UnionSummary, 0, len(unions))
	for _, union := range unions {
		summaries = append(summaries, dto.UnionSummary{
			UnionID:           union.UnionID,
			UnionName:         union.UnionName,
			ChairpersonName:   union.ChairpersonName,
			ChairpersonLevel:  union.ChairpersonLevel,
			UnionLevel:        union.UnionLevel,
			UnionMembers:      union.UnionMembers,
			UnionDesc:         union.UnionDesc,
		})
	}
	
	return &dto.UnionListResponse{
		Unions: summaries,
		Total:  len(summaries),
	}, nil
}

// GetMyUnionRank gets the current user's union rank
func (s *UnionService) GetMyUnionRank(userID int) (*dto.UnionRankResponse, error) {
	// Check if user is in a union
	member, err := s.memberRepo.GetByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("获取用户工会信息失败: %w", err)
	}
	
	if member == nil {
		return &dto.UnionRankResponse{
			InUnion: false,
			Message: "您尚未加入任何工会",
		}, nil
	}
	
	// Get union info
	union, err := s.unionRepo.GetByID(member.UnionID)
	if err != nil {
		return nil, fmt.Errorf("获取工会信息失败: %w", err)
	}
	
	if union == nil {
		// Clean up orphaned member record
		s.memberRepo.DeleteByUserID(userID)
		return &dto.UnionRankResponse{
			InUnion: false,
			Message: "您尚未加入任何工会",
		}, nil
	}
	
	// Get union rank
	rank, err := s.unionRepo.GetUnionRank(member.UnionID)
	if err != nil {
		return nil, fmt.Errorf("获取工会排名失败: %w", err)
	}
	
	return &dto.UnionRankResponse{
		InUnion:   true,
		UnionName: union.UnionName,
		Rank:      rank,
		Message:   fmt.Sprintf("您的工会 %s 当前排名第 %d", union.UnionName, rank),
	}, nil
}