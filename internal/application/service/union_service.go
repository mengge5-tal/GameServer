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
	unionRepo           repository.UnionRepository
	memberRepo          repository.UnionMemberRepository
	requestRepo         repository.UnionRequestRepository
	experienceRepo      repository.UnionExperienceRepository
	playerRepo          repository.PlayerRepository
	userRepo            repository.UserRepository
	cacheService        cache.CacheService
	notificationService NotificationService
	inviteRepo          repository.UnionInviteRepository
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
	notificationService NotificationService,
	inviteRepo repository.UnionInviteRepository,
) *UnionService {
	return &UnionService{
		unionRepo:           unionRepo,
		memberRepo:          memberRepo,
		requestRepo:         requestRepo,
		experienceRepo:      experienceRepo,
		playerRepo:          playerRepo,
		userRepo:            userRepo,
		cacheService:        cacheService,
		notificationService: notificationService,
		inviteRepo:          inviteRepo,
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
		UnionID:          union.UnionID,
		UnionName:        union.UnionName,
		ChairpersonID:    union.ChairpersonID,
		ChairpersonName:  union.ChairpersonName,
		ChairpersonLevel: union.ChairpersonLevel,
		UnionLevel:       union.UnionLevel,
		UnionMembers:     union.UnionMembers,
		Experience:       union.Experience,
		CreatedTime:      union.CreatedTime,
		UnionDesc:        union.UnionDesc,
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
		UnionID:          union.UnionID,
		UnionName:        union.UnionName,
		ChairpersonID:    union.ChairpersonID,
		ChairpersonName:  union.ChairpersonName,
		ChairpersonLevel: union.ChairpersonLevel,
		UnionLevel:       union.UnionLevel,
		UnionMembers:     union.UnionMembers,
		Experience:       union.Experience,
		CreatedTime:      union.CreatedTime,
		UnionDesc:        union.UnionDesc,
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

	// Send real-time notification to chairperson if online
	notification := &dto.UnionJoinRequestNotification{
		RequestID:      request.ID,
		UnionID:        union.UnionID,
		UnionName:      union.UnionName,
		ApplicantID:    req.ApplicantID,
		ApplicantName:  user.Username,
		ApplicantLevel: playerInfo.Level,
		RequestTime:    request.RequestTime.Format("2006-01-02 15:04:05"),
		Message:        fmt.Sprintf("%s (Lv.%d) 申请加入工会 %s", user.Username, playerInfo.Level, union.UnionName),
	}

	// Send notification (ignore if chairperson is offline)
	if err := s.notificationService.SendUnionJoinRequestNotification(union.ChairpersonID, notification); err != nil {
		// Log error but don't fail the join request
		println("Failed to send notification to chairperson:", err.Error())
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
			UnionID:          union.UnionID,
			UnionName:        union.UnionName,
			ChairpersonName:  union.ChairpersonName,
			ChairpersonLevel: union.ChairpersonLevel,
			UnionLevel:       union.UnionLevel,
			UnionMembers:     union.UnionMembers,
			UnionDesc:        union.UnionDesc,
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
		UnionID:          union.UnionID,
		UnionName:        union.UnionName,
		ChairpersonID:    union.ChairpersonID,
		ChairpersonName:  union.ChairpersonName,
		ChairpersonLevel: union.ChairpersonLevel,
		UnionLevel:       union.UnionLevel,
		UnionMembers:     union.UnionMembers,
		Experience:       union.Experience,
		CreatedTime:      union.CreatedTime,
		UnionDesc:        union.UnionDesc,
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
			UnionID:          union.UnionID,
			UnionName:        union.UnionName,
			ChairpersonName:  union.ChairpersonName,
			ChairpersonLevel: union.ChairpersonLevel,
			UnionLevel:       union.UnionLevel,
			UnionMembers:     union.UnionMembers,
			UnionDesc:        union.UnionDesc,
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

// GetUnionRequests gets all union application requests for chairperson
func (s *UnionService) GetUnionRequests(req *dto.GetUnionRequestsRequest) (*dto.UnionRequestListResponse, error) {
	// Verify that the user is a chairperson of some union
	member, err := s.memberRepo.GetByUserID(req.ChairpersonID)
	if err != nil {
		return nil, fmt.Errorf("获取用户工会信息失败: %w", err)
	}

	if member == nil || member.RoleID != entity.UnionRoleLeader {
		return nil, fmt.Errorf("您不是任何工会的会长")
	}

	// Get all pending requests for this union
	requests, err := s.requestRepo.GetByUnionIDAndStatus(member.UnionID, entity.UnionRequestStatusPending)
	if err != nil {
		return nil, fmt.Errorf("获取申请记录失败: %w", err)
	}

	responses := make([]dto.UnionRequestResponse, 0, len(requests))
	for _, request := range requests {
		// Get union info
		union, err := s.unionRepo.GetByID(request.UnionID)
		if err != nil {
			continue // Skip invalid records
		}

		// Get applicant info
		user, err := s.userRepo.GetByID(request.ApplicantID)
		if err != nil {
			continue // Skip invalid records
		}

		player, err := s.playerRepo.GetByUserID(request.ApplicantID)
		applicantLevel := 1
		if err == nil && player != nil {
			applicantLevel = player.Level
		}

		responses = append(responses, dto.UnionRequestResponse{
			ID:             request.ID,
			UnionID:        request.UnionID,
			UnionName:      union.UnionName,
			ApplicantID:    request.ApplicantID,
			ApplicantName:  user.Username,
			ApplicantLevel: applicantLevel,
			RequestStatus:  request.RequestStatus,
			StatusName:     getRequestStatusName(request.RequestStatus),
			RequestTime:    request.RequestTime,
		})
	}

	return &dto.UnionRequestListResponse{
		Requests: responses,
		Total:    len(responses),
	}, nil
}

// getRequestStatusName returns the Chinese name for request status
func getRequestStatusName(status int) string {
	switch status {
	case entity.UnionRequestStatusPending:
		return "待处理"
	case entity.UnionRequestStatusApproved:
		return "已通过"
	case entity.UnionRequestStatusRejected:
		return "已拒绝"
	default:
		return "未知"
	}
}

// InviteToUnion invites a user to join a union
func (s *UnionService) InviteToUnion(req *dto.InviteToUnionRequest) error {
	// Get inviter's union membership
	inviterMember, err := s.memberRepo.GetByUserID(req.InviterID)
	if err != nil {
		return fmt.Errorf("获取邀请人工会信息失败: %w", err)
	}

	if inviterMember == nil {
		return fmt.Errorf("您不在任何工会中，无法邀请他人")
	}

	// Check if inviter has permission to invite (only leaders and vice leaders can invite)
	if inviterMember.RoleID != entity.UnionRoleLeader && inviterMember.RoleID != entity.UnionRoleViceLeader {
		return fmt.Errorf("只有会长和副会长才能邀请成员")
	}

	// Check if invitee is already in a union
	isInUnion, err := s.memberRepo.IsUserInUnion(req.InviteToUserID)
	if err != nil {
		return fmt.Errorf("检查被邀请人工会状态失败: %w", err)
	}

	if isInUnion {
		return fmt.Errorf("该玩家已经加入了工会")
	}

	// Get inviter and invitee info
	inviter, err := s.userRepo.GetByID(req.InviterID)
	if err != nil {
		return fmt.Errorf("获取邀请人信息失败: %w", err)
	}

	if inviter == nil {
		return fmt.Errorf("邀请人不存在")
	}

	invitee, err := s.userRepo.GetByID(req.InviteToUserID)
	if err != nil {
		return fmt.Errorf("获取被邀请人信息失败: %w", err)
	}

	if invitee == nil {
		return fmt.Errorf("被邀请人不存在")
	}

	// Get union info
	union, err := s.unionRepo.GetByID(inviterMember.UnionID)
	if err != nil {
		return fmt.Errorf("获取工会信息失败: %w", err)
	}

	if union == nil {
		return fmt.Errorf("工会不存在")
	}

	// Check if there's already a pending invite
	hasPending, err := s.inviteRepo.HasPendingInvite(inviter.Username, invitee.Username, union.UnionID)
	if err != nil {
		return fmt.Errorf("检查待处理邀请失败: %w", err)
	}

	if hasPending {
		return fmt.Errorf("您已经向该玩家发送了邀请，请等待处理")
	}

	// Note: inviter player info not needed for basic invite creation

	// Create union invite
	invite := &entity.UnionInvite{
		InviteFromUser:   inviter.Username,
		InviteToUser:     invitee.Username,
		UnionID:          union.UnionID,
		UnionName:        union.UnionName,
		ChairpersonID:    union.ChairpersonID,
		ChairpersonName:  union.ChairpersonName,
		ChairpersonLevel: union.ChairpersonLevel,
		UnionLevel:       union.UnionLevel,
		CreateTime:       time.Now(),
		Status:           entity.UnionInviteStatusPending,
	}

	err = s.inviteRepo.Create(invite)
	if err != nil {
		return fmt.Errorf("创建工会邀请失败: %w", err)
	}

	// Send real-time notification to invitee if online
	notification := &dto.UnionInviteNotification{
		InviteID:         invite.ID,
		UnionID:          union.UnionID,
		UnionName:        union.UnionName,
		InviterName:      inviter.Username,
		ChairpersonName:  union.ChairpersonName,
		ChairpersonLevel: union.ChairpersonLevel,
		UnionLevel:       union.UnionLevel,
		CreateTime:       invite.CreateTime.Format("2006-01-02 15:04:05"),
		Message:          fmt.Sprintf("%s 邀请您加入工会 %s", inviter.Username, union.UnionName),
	}

	// Send notification (ignore if user is offline)
	if err := s.notificationService.SendUnionInviteNotification(req.InviteToUserID, notification); err != nil {
		// Log error but don't fail the invite
		println("Failed to send invite notification to user:", err.Error())
	}

	return nil
}

// GetUnionInvites gets all union invitations for a user
func (s *UnionService) GetUnionInvites(req *dto.GetUnionInvitesRequest) (*dto.UnionInviteListResponse, error) {
	invites, err := s.inviteRepo.GetPendingByUserID(req.UserID)
	if err != nil {
		return nil, fmt.Errorf("获取工会邀请失败: %w", err)
	}

	responses := make([]dto.UnionInviteResponse, 0, len(invites))
	for _, invite := range invites {
		responses = append(responses, dto.UnionInviteResponse{
			ID:               invite.ID,
			InviteFromUser:   invite.InviteFromUser,
			InviteToUser:     invite.InviteToUser,
			UnionID:          invite.UnionID,
			UnionName:        invite.UnionName,
			ChairpersonID:    invite.ChairpersonID,
			ChairpersonName:  invite.ChairpersonName,
			ChairpersonLevel: invite.ChairpersonLevel,
			UnionLevel:       invite.UnionLevel,
			CreateTime:       invite.CreateTime.Format("2006-01-02 15:04:05"),
			Status:           invite.Status,
			StatusName:       invite.GetStatusName(),
		})
	}

	return &dto.UnionInviteListResponse{
		Invites: responses,
		Total:   len(responses),
	}, nil
}

// ProcessUnionInvite processes a union invitation (accept/reject)
func (s *UnionService) ProcessUnionInvite(req *dto.ProcessUnionInviteRequest) error {
	// Get the invite details
	invite, err := s.inviteRepo.GetByID(req.InviteID)
	if err != nil {
		return fmt.Errorf("获取邀请信息失败: %w", err)
	}

	if invite == nil {
		return fmt.Errorf("邀请不存在")
	}

	// Verify that the user is the invitee
	user, err := s.userRepo.GetByID(req.UserID)
	if err != nil {
		return fmt.Errorf("获取用户信息失败: %w", err)
	}

	if user.Username != invite.InviteToUser {
		return fmt.Errorf("您不是该邀请的接收者")
	}

	// Check if invite is still pending
	if !invite.IsPending() {
		return fmt.Errorf("该邀请已被处理")
	}

	// Check if user is already in a union
	isInUnion, err := s.memberRepo.IsUserInUnion(req.UserID)
	if err != nil {
		return fmt.Errorf("检查用户工会状态失败: %w", err)
	}

	if isInUnion {
		// Update invite status to rejected since user is already in a union
		s.inviteRepo.ProcessInvite(req.InviteID, entity.UnionInviteStatusRejected)
		return fmt.Errorf("您已经加入了工会")
	}

	// If accepting the invite
	if req.Status == entity.UnionInviteStatusAccepted {
		// Get player info
		playerInfo, err := s.playerRepo.GetByUserID(req.UserID)
		if err != nil {
			return fmt.Errorf("获取玩家信息失败: %w", err)
		}

		// Add member to union
		member := &entity.UnionMember{
			UnionID:     invite.UnionID,
			UnionName:   invite.UnionName,
			MemberID:    req.UserID,
			MemberLevel: playerInfo.Level,
			JoinedTime:  time.Now(),
			RoleID:      entity.UnionRoleMember,
		}

		err = s.memberRepo.Create(member)
		if err != nil {
			return fmt.Errorf("添加工会成员失败: %w", err)
		}

		// Update union member count
		err = s.unionRepo.IncrementMemberCount(invite.UnionID)
		if err != nil {
			// Rollback member creation
			s.memberRepo.DeleteByUserID(req.UserID)
			return fmt.Errorf("更新工会成员数量失败: %w", err)
		}

		// Clear cache
		s.cacheService.Delete(fmt.Sprintf("user_union:%d", req.UserID))
	}

	// Update invite status
	err = s.inviteRepo.ProcessInvite(req.InviteID, req.Status)
	if err != nil {
		return fmt.Errorf("更新邀请状态失败: %w", err)
	}

	return nil
}

// PromoteMember promotes a member to vice leader
func (s *UnionService) PromoteMember(req *dto.PromoteMemberRequest) error {
	// Get leader's union membership and verify permissions
	leaderMember, err := s.memberRepo.GetByUserID(req.LeaderID)
	if err != nil {
		return fmt.Errorf("获取会长信息失败: %w", err)
	}

	if leaderMember == nil || leaderMember.RoleID != entity.UnionRoleLeader {
		return fmt.Errorf("只有会长才能提升成员权限")
	}

	// Get member to be promoted
	member, err := s.memberRepo.GetByUserID(req.MemberUserID)
	if err != nil {
		return fmt.Errorf("获取成员信息失败: %w", err)
	}

	if member == nil {
		return fmt.Errorf("成员不存在")
	}

	// Check if they are in the same union
	if member.UnionID != leaderMember.UnionID {
		return fmt.Errorf("只能提升本工会成员")
	}

	// Check current role
	if member.RoleID != entity.UnionRoleMember {
		return fmt.Errorf("该成员已经是副会长或会长")
	}

	// Update role to vice leader
	err = s.memberRepo.UpdateRole(req.MemberUserID, member.UnionID, entity.UnionRoleViceLeader)
	if err != nil {
		return fmt.Errorf("提升成员权限失败: %w", err)
	}

	// Clear cache
	s.cacheService.Delete(fmt.Sprintf("user_union:%d", req.MemberUserID))

	return nil
}

// DemoteMember demotes a vice leader to regular member
func (s *UnionService) DemoteMember(req *dto.DemoteMemberRequest) error {
	// Get leader's union membership and verify permissions
	leaderMember, err := s.memberRepo.GetByUserID(req.LeaderID)
	if err != nil {
		return fmt.Errorf("获取会长信息失败: %w", err)
	}

	if leaderMember == nil || leaderMember.RoleID != entity.UnionRoleLeader {
		return fmt.Errorf("只有会长才能降级成员权限")
	}

	// Get member to be demoted
	member, err := s.memberRepo.GetByUserID(req.MemberUserID)
	if err != nil {
		return fmt.Errorf("获取成员信息失败: %w", err)
	}

	if member == nil {
		return fmt.Errorf("成员不存在")
	}

	// Check if they are in the same union
	if member.UnionID != leaderMember.UnionID {
		return fmt.Errorf("只能降级本工会成员")
	}

	// Check current role
	if member.RoleID != entity.UnionRoleViceLeader {
		return fmt.Errorf("该成员不是副会长")
	}

	// Update role to regular member
	err = s.memberRepo.UpdateRole(req.MemberUserID, member.UnionID, entity.UnionRoleMember)
	if err != nil {
		return fmt.Errorf("降级成员权限失败: %w", err)
	}

	// Clear cache
	s.cacheService.Delete(fmt.Sprintf("user_union:%d", req.MemberUserID))

	return nil
}

// KickMember kicks a member from the union
func (s *UnionService) KickMember(req *dto.KickMemberRequest) error {
	// Get kicker's union membership and verify permissions
	kickerMember, err := s.memberRepo.GetByUserID(req.KickerID)
	if err != nil {
		return fmt.Errorf("获取操作者信息失败: %w", err)
	}

	if kickerMember == nil {
		return fmt.Errorf("操作者不在任何工会中")
	}

	// Check permissions: leader and vice leader can kick regular members, leader can kick vice leaders
	if kickerMember.RoleID != entity.UnionRoleLeader && kickerMember.RoleID != entity.UnionRoleViceLeader {
		return fmt.Errorf("只有会长和副会长才能踢出成员")
	}

	// Get member to be kicked
	member, err := s.memberRepo.GetByUserID(req.MemberUserID)
	if err != nil {
		return fmt.Errorf("获取被踢出成员信息失败: %w", err)
	}

	if member == nil {
		return fmt.Errorf("被踢出成员不存在")
	}

	// Check if they are in the same union
	if member.UnionID != kickerMember.UnionID {
		return fmt.Errorf("只能踢出本工会成员")
	}

	// Check role permissions
	if member.RoleID == entity.UnionRoleLeader {
		return fmt.Errorf("不能踢出会长")
	}

	if member.RoleID == entity.UnionRoleViceLeader && kickerMember.RoleID != entity.UnionRoleLeader {
		return fmt.Errorf("只有会长才能踢出副会长")
	}

	// Cannot kick yourself
	if req.MemberUserID == req.KickerID {
		return fmt.Errorf("不能踢出自己")
	}

	// Remove member from union
	err = s.memberRepo.DeleteByUserID(req.MemberUserID)
	if err != nil {
		return fmt.Errorf("踢出成员失败: %w", err)
	}

	// Update union member count
	err = s.unionRepo.DecrementMemberCount(member.UnionID)
	if err != nil {
		// This is not critical, log but don't fail
		fmt.Printf("Warning: Failed to update union member count: %v", err)
	}

	// Clear cache
	s.cacheService.Delete(fmt.Sprintf("user_union:%d", req.MemberUserID))

	return nil
}

// TransferLeadership transfers leadership to another member
func (s *UnionService) TransferLeadership(req *dto.TransferLeadershipRequest) error {
	// Get current leader's union membership and verify permissions
	currentLeader, err := s.memberRepo.GetByUserID(req.CurrentLeaderID)
	if err != nil {
		return fmt.Errorf("获取当前会长信息失败: %w", err)
	}

	if currentLeader == nil || currentLeader.RoleID != entity.UnionRoleLeader {
		return fmt.Errorf("只有会长才能转让会长职位")
	}

	// Get new leader
	newLeader, err := s.memberRepo.GetByUserID(req.NewLeaderUserID)
	if err != nil {
		return fmt.Errorf("获取新会长信息失败: %w", err)
	}

	if newLeader == nil {
		return fmt.Errorf("新会长不存在")
	}

	// Check if they are in the same union
	if newLeader.UnionID != currentLeader.UnionID {
		return fmt.Errorf("只能将会长职位转让给本工会成员")
	}

	// Cannot transfer to yourself
	if req.NewLeaderUserID == req.CurrentLeaderID {
		return fmt.Errorf("不能将会长职位转让给自己")
	}

	// Determine current leader's new role based on new leader's current role
	var currentLeaderNewRole int
	if newLeader.RoleID == entity.UnionRoleViceLeader {
		currentLeaderNewRole = entity.UnionRoleViceLeader
	} else {
		currentLeaderNewRole = entity.UnionRoleMember
	}

	// Update new leader to leader role
	err = s.memberRepo.UpdateRole(req.NewLeaderUserID, newLeader.UnionID, entity.UnionRoleLeader)
	if err != nil {
		return fmt.Errorf("提升新会长失败: %w", err)
	}

	// Update current leader to new role
	err = s.memberRepo.UpdateRole(req.CurrentLeaderID, currentLeader.UnionID, currentLeaderNewRole)
	if err != nil {
		// Rollback new leader promotion
		s.memberRepo.UpdateRole(req.NewLeaderUserID, newLeader.UnionID, newLeader.RoleID)
		return fmt.Errorf("降级当前会长失败: %w", err)
	}

	// Get union info and update chairperson information
	union, err := s.unionRepo.GetByID(currentLeader.UnionID)
	if err != nil {
		// Log error but don't rollback as the role transfer is already complete
		fmt.Printf("Warning: Failed to get union info for leadership transfer: %v", err)
		return nil
	}

	// Get new leader's user info
	newLeaderUser, err := s.userRepo.GetByID(req.NewLeaderUserID)
	if err != nil {
		fmt.Printf("Warning: Failed to get new leader user info: %v", err)
		return nil
	}

	// Get new leader's player info for level
	newLeaderPlayer, err := s.playerRepo.GetByUserID(req.NewLeaderUserID)
	if err != nil {
		fmt.Printf("Warning: Failed to get new leader player info: %v", err)
		return nil
	}

	// Update union chairperson info
	union.ChairpersonID = req.NewLeaderUserID
	union.ChairpersonName = newLeaderUser.Username
	union.ChairpersonLevel = newLeaderPlayer.Level

	err = s.unionRepo.Update(union)
	if err != nil {
		fmt.Printf("Warning: Failed to update union chairperson info: %v", err)
	}

	// Clear cache for both users
	s.cacheService.Delete(fmt.Sprintf("user_union:%d", req.CurrentLeaderID))
	s.cacheService.Delete(fmt.Sprintf("user_union:%d", req.NewLeaderUserID))

	return nil
}

// GetUnionMembers gets union member list with pagination
func (s *UnionService) GetUnionMembers(req *dto.GetUnionMembersRequest) (*dto.UnionMemberListResponse, error) {
	// Validate union ID
	if req.UnionID <= 0 {
		return nil, fmt.Errorf("工会ID无效")
	}

	// Set default pagination values
	page := req.Page
	if page <= 0 {
		page = 1
	}
	limit := req.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	// Verify union exists
	union, err := s.unionRepo.GetByID(req.UnionID)
	if err != nil {
		return nil, fmt.Errorf("获取工会信息失败: %w", err)
	}
	if union == nil {
		return nil, fmt.Errorf("工会不存在")
	}

	// Get union members with pagination
	members, total, err := s.memberRepo.GetMembersByUnionIDWithPagination(req.UnionID, page, limit)
	if err != nil {
		return nil, fmt.Errorf("获取工会成员列表失败: %w", err)
	}

	// Convert to response format
	memberResponses := make([]dto.UnionMemberResponse, 0, len(members))
	for _, member := range members {
		// Get user online status (you can implement this based on your online user tracking)
		isOnline := false // TODO: Implement online status check
		
		// Determine role name
		roleName := getRoleName(member.RoleID)
		
		memberResponses = append(memberResponses, dto.UnionMemberResponse{
			UserID:       member.MemberID,
			Username:     member.MemberName,
			Level:        member.MemberLevel,
			RoleID:       member.RoleID,
			RoleName:     roleName,
			JoinTime:     member.JoinedTime.Format("2006-01-02 15:04:05"),
			LastLogin:    member.LastLogin,
			Experience:   member.UserExperience,
			IsOnline:     isOnline,
		})
	}

	// Calculate total pages
	totalPages := (total + limit - 1) / limit

	return &dto.UnionMemberListResponse{
		Members:    memberResponses,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

// SearchUnionMembers searches union members by keyword
func (s *UnionService) SearchUnionMembers(req *dto.SearchUnionMembersRequest) (*dto.SearchUnionMembersResponse, error) {
	// Validate union ID
	if req.UnionID <= 0 {
		return nil, fmt.Errorf("工会ID无效")
	}

	// Validate keyword
	if len(req.Keyword) == 0 {
		return nil, fmt.Errorf("搜索关键字不能为空")
	}

	// Set default pagination values
	page := req.Page
	if page <= 0 {
		page = 1
	}
	limit := req.Limit
	if limit <= 0 || limit > 50 {
		limit = 10 // 搜索结果默认每页10条，最多50条
	}

	// Verify union exists
	union, err := s.unionRepo.GetByID(req.UnionID)
	if err != nil {
		return nil, fmt.Errorf("获取工会信息失败: %w", err)
	}
	if union == nil {
		return nil, fmt.Errorf("工会不存在")
	}

	// Search union members with pagination
	members, total, err := s.memberRepo.SearchMembersByUnionIDAndKeyword(req.UnionID, req.Keyword, page, limit)
	if err != nil {
		return nil, fmt.Errorf("搜索工会成员失败: %w", err)
	}

	// Convert to response format
	memberResponses := make([]dto.UnionMemberResponse, 0, len(members))
	for _, member := range members {
		// Get user online status (you can implement this based on your online user tracking)
		isOnline := false // TODO: Implement online status check
		
		// Determine role name
		roleName := getRoleName(member.RoleID)
		
		memberResponses = append(memberResponses, dto.UnionMemberResponse{
			UserID:       member.MemberID,
			Username:     member.MemberName,
			Level:        member.MemberLevel,
			RoleID:       member.RoleID,
			RoleName:     roleName,
			JoinTime:     member.JoinedTime.Format("2006-01-02 15:04:05"),
			LastLogin:    member.LastLogin,
			Experience:   member.UserExperience,
			IsOnline:     isOnline,
		})
	}

	// Calculate total pages
	totalPages := (total + limit - 1) / limit

	return &dto.SearchUnionMembersResponse{
		Members:    memberResponses,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
		Keyword:    req.Keyword,
	}, nil
}

// UpdateUnionInfo updates union information (chairman only)
func (s *UnionService) UpdateUnionInfo(req *dto.UpdateUnionInfoRequest) (*dto.UnionResponse, error) {
	// Validate input
	if req.ChairpersonID <= 0 {
		return nil, fmt.Errorf("会长ID无效")
	}
	
	// Validate union name
	if len(req.UnionName) < 2 || len(req.UnionName) > 20 {
		return nil, fmt.Errorf("工会名称长度必须在2-20个字符之间")
	}
	
	// Validate description (optional but if provided should be reasonable length)
	if len(req.Description) > 200 {
		return nil, fmt.Errorf("工会简介长度不能超过200个字符")
	}
	
	// Get current user's union membership
	currentMember, err := s.memberRepo.GetByUserID(req.ChairpersonID)
	if err != nil {
		return nil, fmt.Errorf("获取用户工会信息失败: %w", err)
	}
	
	if currentMember == nil {
		return nil, fmt.Errorf("您不在任何工会中")
	}
	
	// Check if user is the leader (chairman)
	if currentMember.RoleID != entity.UnionRoleLeader {
		return nil, fmt.Errorf("只有工会会长才能修改工会信息")
	}
	
	// Check if new union name already exists (only if different from current)
	currentUnion, err := s.unionRepo.GetByID(currentMember.UnionID)
	if err != nil {
		return nil, fmt.Errorf("获取当前工会信息失败: %w", err)
	}
	
	if currentUnion == nil {
		return nil, fmt.Errorf("工会不存在")
	}
	
	// Only check name uniqueness if the name is being changed
	if req.UnionName != currentUnion.UnionName {
		exists, err := s.unionRepo.Exists(req.UnionName)
		if err != nil {
			return nil, fmt.Errorf("检查工会名称失败: %w", err)
		}
		if exists {
			return nil, fmt.Errorf("工会名称已存在")
		}
	}
	
	// Update union information in database
	err = s.unionRepo.UpdateUnionInfo(currentMember.UnionID, req.UnionName, req.Description)
	if err != nil {
		return nil, fmt.Errorf("更新工会信息失败: %w", err)
	}
	
	// Update union member records with new union name if changed
	if req.UnionName != currentUnion.UnionName {
		// Get all union members to update their union name
		allMembers, err := s.memberRepo.GetByUnionID(currentMember.UnionID)
		if err != nil {
			return nil, fmt.Errorf("获取工会成员列表失败: %w", err)
		}
		
		// Update each member's union name
		for _, member := range allMembers {
			member.UnionName = req.UnionName
			err = s.memberRepo.Update(member)
			if err != nil {
				return nil, fmt.Errorf("更新成员工会名称失败: %w", err)
			}
		}
	}
	
	// Clear cache
	s.cacheService.Delete(fmt.Sprintf("union:%d", currentMember.UnionID))
	
	// Get updated union information to return
	updatedUnion, err := s.unionRepo.GetByID(currentMember.UnionID)
	if err != nil {
		return nil, fmt.Errorf("获取更新后的工会信息失败: %w", err)
	}
	
	return &dto.UnionResponse{
		UnionID:           updatedUnion.UnionID,
		UnionName:         updatedUnion.UnionName,
		ChairpersonID:     updatedUnion.ChairpersonID,
		ChairpersonName:   updatedUnion.ChairpersonName,
		ChairpersonLevel:  updatedUnion.ChairpersonLevel,
		UnionLevel:        updatedUnion.UnionLevel,
		UnionMembers:      updatedUnion.UnionMembers,
		Experience:        updatedUnion.Experience,
		CreatedTime:       updatedUnion.CreatedTime,
		UnionDesc:         updatedUnion.UnionDesc,
	}, nil
}

// getRoleName converts role ID to role name
func getRoleName(roleID int) string {
	switch roleID {
	case entity.UnionRoleLeader:
		return "leader"
	case entity.UnionRoleViceLeader:
		return "vice_leader"
	case entity.UnionRoleMember:
		return "member"
	default:
		return "member"
	}
}
