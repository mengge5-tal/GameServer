package websocket

import (
	"GameServer/internal/application/dto"
	"GameServer/internal/domain/valueobject"
	"encoding/json"
	"log"
)

// UnionHandler handles union-related messages
type UnionHandler struct {
	unionService UnionServiceInterface
}

// NewUnionHandler creates a new union handler
func NewUnionHandler(unionService UnionServiceInterface) *UnionHandler {
	return &UnionHandler{unionService: unionService}
}

// Handle handles union messages
func (h *UnionHandler) Handle(client *Client, message *valueobject.Message) *valueobject.Response {
	switch message.Action {
	case valueobject.ActionGetMyUnion:
		return h.handleGetMyUnion(client, message)
	case valueobject.ActionCreateUnion:
		return h.handleCreateUnion(client, message)
	case valueobject.ActionJoinUnion:
		return h.handleJoinUnion(client, message)
	case valueobject.ActionGetRecommendedUnions:
		return h.handleGetRecommendedUnions(client, message)
	case valueobject.ActionProcessUnionRequest:
		return h.handleProcessUnionRequest(client, message)
	case valueobject.ActionGetUnionInfo:
		return h.handleGetUnionInfo(client, message)
	case valueobject.ActionLeaveUnion:
		return h.handleLeaveUnion(client, message)
	case valueobject.ActionDismissUnion:
		return h.handleDismissUnion(client, message)
	case valueobject.ActionGetUnionRanking:
		return h.handleGetUnionRanking(client, message)
	case valueobject.ActionGetMyUnionRank:
		return h.handleGetMyUnionRank(client, message)
	case valueobject.ActionGetUnionRequests:
		return h.handleGetUnionRequests(client, message)
	case valueobject.ActionInviteToUnion:
		return h.handleInviteToUnion(client, message)
	case valueobject.ActionGetUnionInvites:
		return h.handleGetUnionInvites(client, message)
	case valueobject.ActionProcessUnionInvite:
		return h.handleProcessUnionInvite(client, message)
	case valueobject.ActionPromoteMember:
		return h.handlePromoteMember(client, message)
	case valueobject.ActionDemoteMember:
		return h.handleDemoteMember(client, message)
	case valueobject.ActionKickMember:
		return h.handleKickMember(client, message)
	case valueobject.ActionTransferLeadership:
		return h.handleTransferLeadership(client, message)
	case valueobject.ActionGetUnionMembers:
		return h.handleGetUnionMembers(client, message)
	default:
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInvalidRequest, "Unknown union action")
	}
}

// handleGetMyUnion handles getting user's current union information
func (h *UnionHandler) handleGetMyUnion(client *Client, message *valueobject.Message) *valueobject.Response {
	userID := client.GetUserID()
	if userID == 0 {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeUnauthorized, "User not authenticated")
	}

	unionInfo, err := h.unionService.GetMyUnionInfo(userID)
	if err != nil {
		log.Printf("Get my union error: %v", err)
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInternalError, "获取工会信息失败")
	}

	if unionInfo == nil {
		// User is not in any union
		return valueobject.NewSuccessResponseWithUniqueID(message.Type, message.Action, map[string]interface{}{
			"in_union": false,
			"message":  "您尚未加入任何工会",
		})
	}

	return valueobject.NewSuccessResponseWithUniqueID(message.Type, message.Action, map[string]interface{}{
		"in_union":   true,
		"union_info": unionInfo,
	})
}

// handleCreateUnion handles union creation
func (h *UnionHandler) handleCreateUnion(client *Client, message *valueobject.Message) *valueobject.Response {
	userID := client.GetUserID()
	if userID == 0 {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeUnauthorized, "User not authenticated")
	}

	var req dto.CreateUnionRequest
	if err := json.Unmarshal(message.Data, &req); err != nil {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInvalidRequest, "Invalid union data")
	}

	// Set chairperson ID from authenticated user
	req.ChairpersonID = userID

	unionInfo, err := h.unionService.CreateUnion(&req)
	if err != nil {
		log.Printf("Create union error: %v", err)
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeValidationError, err.Error())
	}

	return valueobject.NewSuccessResponseWithUniqueID(message.Type, message.Action, unionInfo)
}

// handleJoinUnion handles joining a union (creates application)
func (h *UnionHandler) handleJoinUnion(client *Client, message *valueobject.Message) *valueobject.Response {
	userID := client.GetUserID()
	if userID == 0 {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeUnauthorized, "User not authenticated")
	}

	var req dto.JoinUnionRequest
	if err := json.Unmarshal(message.Data, &req); err != nil {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInvalidRequest, "Invalid join union data")
	}

	// Set applicant ID from authenticated user
	req.ApplicantID = userID

	err := h.unionService.JoinUnion(&req)
	if err != nil {
		log.Printf("Join union error: %v", err)
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeValidationError, err.Error())
	}

	return valueobject.NewSuccessResponseWithUniqueID(message.Type, message.Action, map[string]interface{}{
		"message": "加入申请已提交，请等待工会会长审核",
	})
}

// handleGetRecommendedUnions handles getting recommended unions
func (h *UnionHandler) handleGetRecommendedUnions(client *Client, message *valueobject.Message) *valueobject.Response {
	// Parse limit from message data (optional)
	var req struct {
		Limit int `json:"limit,omitempty"`
	}
	
	// Default limit is 6, but allow client to override
	limit := 6
	if len(message.Data) > 0 {
		if err := json.Unmarshal(message.Data, &req); err == nil && req.Limit > 0 {
			limit = req.Limit
		}
	}

	unions, err := h.unionService.GetRecommendedUnions(limit)
	if err != nil {
		log.Printf("Get recommended unions error: %v", err)
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInternalError, "获取推荐工会失败")
	}

	return valueobject.NewSuccessResponseWithUniqueID(message.Type, message.Action, unions)
}

// handleProcessUnionRequest handles processing union join requests (approve/reject)
func (h *UnionHandler) handleProcessUnionRequest(client *Client, message *valueobject.Message) *valueobject.Response {
	userID := client.GetUserID()
	if userID == 0 {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeUnauthorized, "User not authenticated")
	}

	var req dto.ProcessUnionRequestDTO
	if err := json.Unmarshal(message.Data, &req); err != nil {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInvalidRequest, "Invalid process request data")
	}

	// Set chairperson ID from authenticated user
	req.ChairpersonID = userID

	err := h.unionService.ProcessUnionRequest(&req)
	if err != nil {
		log.Printf("Process union request error: %v", err)
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeValidationError, err.Error())
	}

	statusText := "已拒绝"
	if req.Status == 2 { // approved
		statusText = "已通过"
	}

	return valueobject.NewSuccessResponseWithUniqueID(message.Type, message.Action, map[string]interface{}{
		"message": "申请处理成功：" + statusText,
	})
}

// handleGetUnionInfo handles getting detailed union information
func (h *UnionHandler) handleGetUnionInfo(client *Client, message *valueobject.Message) *valueobject.Response {
	var req struct {
		UnionID int `json:"union_id"`
	}
	
	if err := json.Unmarshal(message.Data, &req); err != nil {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInvalidRequest, "Invalid union info request data")
	}

	unionInfo, err := h.unionService.GetUnionInfo(req.UnionID)
	if err != nil {
		log.Printf("Get union info error: %v", err)
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInternalError, err.Error())
	}

	if unionInfo == nil {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeNotFound, "工会不存在")
	}

	return valueobject.NewSuccessResponseWithUniqueID(message.Type, message.Action, unionInfo)
}

// handleLeaveUnion handles leaving a union
func (h *UnionHandler) handleLeaveUnion(client *Client, message *valueobject.Message) *valueobject.Response {
	userID := client.GetUserID()
	if userID == 0 {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeUnauthorized, "User not authenticated")
	}

	err := h.unionService.LeaveUnion(userID)
	if err != nil {
		log.Printf("Leave union error: %v", err)
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeValidationError, err.Error())
	}

	return valueobject.NewSuccessResponseWithUniqueID(message.Type, message.Action, map[string]interface{}{
		"message": "成功退出工会",
	})
}

// handleDismissUnion handles dismissing/dissolving a union
func (h *UnionHandler) handleDismissUnion(client *Client, message *valueobject.Message) *valueobject.Response {
	userID := client.GetUserID()
	if userID == 0 {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeUnauthorized, "User not authenticated")
	}

	var req dto.DismissUnionRequest
	if err := json.Unmarshal(message.Data, &req); err != nil {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInvalidRequest, "Invalid dismiss union data")
	}

	// Set chairperson ID from authenticated user
	req.ChairpersonID = userID

	err := h.unionService.DismissUnion(&req)
	if err != nil {
		log.Printf("Dismiss union error: %v", err)
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeValidationError, err.Error())
	}

	return valueobject.NewSuccessResponseWithUniqueID(message.Type, message.Action, map[string]interface{}{
		"message": "工会已解散",
	})
}

// handleGetUnionRanking handles getting union ranking
func (h *UnionHandler) handleGetUnionRanking(client *Client, message *valueobject.Message) *valueobject.Response {
	var req struct {
		Limit int `json:"limit,omitempty"`
	}
	
	// Default limit is 10
	limit := 10
	if len(message.Data) > 0 {
		if err := json.Unmarshal(message.Data, &req); err == nil && req.Limit > 0 {
			limit = req.Limit
		}
	}

	ranking, err := h.unionService.GetUnionRanking(limit)
	if err != nil {
		log.Printf("Get union ranking error: %v", err)
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInternalError, "获取工会排行榜失败")
	}

	return valueobject.NewSuccessResponseWithUniqueID(message.Type, message.Action, ranking)
}

// handleGetMyUnionRank handles getting current user's union rank
func (h *UnionHandler) handleGetMyUnionRank(client *Client, message *valueobject.Message) *valueobject.Response {
	userID := client.GetUserID()
	if userID == 0 {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeUnauthorized, "User not authenticated")
	}

	rank, err := h.unionService.GetMyUnionRank(userID)
	if err != nil {
		log.Printf("Get my union rank error: %v", err)
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInternalError, err.Error())
	}

	return valueobject.NewSuccessResponseWithUniqueID(message.Type, message.Action, rank)
}

// handleGetUnionRequests handles getting union application requests for chairperson
func (h *UnionHandler) handleGetUnionRequests(client *Client, message *valueobject.Message) *valueobject.Response {
	userID := client.GetUserID()
	if userID == 0 {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeUnauthorized, "User not authenticated")
	}

	req := &dto.GetUnionRequestsRequest{
		ChairpersonID: userID,
	}

	requests, err := h.unionService.GetUnionRequests(req)
	if err != nil {
		log.Printf("Get union requests error: %v", err)
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInternalError, err.Error())
	}

	return valueobject.NewSuccessResponseWithUniqueID(message.Type, message.Action, requests)
}

// handleInviteToUnion handles inviting a user to join a union
func (h *UnionHandler) handleInviteToUnion(client *Client, message *valueobject.Message) *valueobject.Response {
	userID := client.GetUserID()
	if userID == 0 {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeUnauthorized, "User not authenticated")
	}

	var req dto.InviteToUnionRequest
	if err := json.Unmarshal(message.Data, &req); err != nil {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInvalidRequest, "Invalid invite data")
	}

	// Set inviter ID from authenticated user
	req.InviterID = userID

	err := h.unionService.InviteToUnion(&req)
	if err != nil {
		log.Printf("Invite to union error: %v", err)
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeValidationError, err.Error())
	}

	return valueobject.NewSuccessResponseWithUniqueID(message.Type, message.Action, map[string]interface{}{
		"message": "工会邀请已发送",
	})
}

// handleGetUnionInvites handles getting user's union invitations
func (h *UnionHandler) handleGetUnionInvites(client *Client, message *valueobject.Message) *valueobject.Response {
	userID := client.GetUserID()
	if userID == 0 {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeUnauthorized, "User not authenticated")
	}

	req := &dto.GetUnionInvitesRequest{
		UserID: userID,
	}

	invites, err := h.unionService.GetUnionInvites(req)
	if err != nil {
		log.Printf("Get union invites error: %v", err)
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInternalError, err.Error())
	}

	return valueobject.NewSuccessResponseWithUniqueID(message.Type, message.Action, invites)
}

// handleProcessUnionInvite handles processing a union invitation (accept/reject)
func (h *UnionHandler) handleProcessUnionInvite(client *Client, message *valueobject.Message) *valueobject.Response {
	userID := client.GetUserID()
	if userID == 0 {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeUnauthorized, "User not authenticated")
	}

	var req dto.ProcessUnionInviteRequest
	if err := json.Unmarshal(message.Data, &req); err != nil {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInvalidRequest, "Invalid process invite data")
	}

	// Set user ID from authenticated user
	req.UserID = userID

	err := h.unionService.ProcessUnionInvite(&req)
	if err != nil {
		log.Printf("Process union invite error: %v", err)
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeValidationError, err.Error())
	}

	statusText := "已拒绝"
	if req.Status == "accepted" {
		statusText = "已接受"
	}

	return valueobject.NewSuccessResponseWithUniqueID(message.Type, message.Action, map[string]interface{}{
		"message": "工会邀请处理成功：" + statusText,
	})
}

// handlePromoteMember handles promoting a member to vice leader
func (h *UnionHandler) handlePromoteMember(client *Client, message *valueobject.Message) *valueobject.Response {
	userID := client.GetUserID()
	if userID == 0 {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeUnauthorized, "User not authenticated")
	}

	var req dto.PromoteMemberRequest
	if err := json.Unmarshal(message.Data, &req); err != nil {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInvalidRequest, "Invalid promote member data")
	}

	// Set leader ID from authenticated user
	req.LeaderID = userID

	err := h.unionService.PromoteMember(&req)
	if err != nil {
		log.Printf("Promote member error: %v", err)
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeValidationError, err.Error())
	}

	return valueobject.NewSuccessResponseWithUniqueID(message.Type, message.Action, map[string]interface{}{
		"message": "成员提升为副会长成功",
	})
}

// handleDemoteMember handles demoting a vice leader to regular member
func (h *UnionHandler) handleDemoteMember(client *Client, message *valueobject.Message) *valueobject.Response {
	userID := client.GetUserID()
	if userID == 0 {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeUnauthorized, "User not authenticated")
	}

	var req dto.DemoteMemberRequest
	if err := json.Unmarshal(message.Data, &req); err != nil {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInvalidRequest, "Invalid demote member data")
	}

	// Set leader ID from authenticated user
	req.LeaderID = userID

	err := h.unionService.DemoteMember(&req)
	if err != nil {
		log.Printf("Demote member error: %v", err)
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeValidationError, err.Error())
	}

	return valueobject.NewSuccessResponseWithUniqueID(message.Type, message.Action, map[string]interface{}{
		"message": "副会长降级为普通成员成功",
	})
}

// handleKickMember handles kicking a member from the union
func (h *UnionHandler) handleKickMember(client *Client, message *valueobject.Message) *valueobject.Response {
	userID := client.GetUserID()
	if userID == 0 {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeUnauthorized, "User not authenticated")
	}

	var req dto.KickMemberRequest
	if err := json.Unmarshal(message.Data, &req); err != nil {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInvalidRequest, "Invalid kick member data")
	}

	// Set kicker ID from authenticated user
	req.KickerID = userID

	err := h.unionService.KickMember(&req)
	if err != nil {
		log.Printf("Kick member error: %v", err)
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeValidationError, err.Error())
	}

	return valueobject.NewSuccessResponseWithUniqueID(message.Type, message.Action, map[string]interface{}{
		"message": "踢出成员成功",
	})
}

// handleTransferLeadership handles transferring leadership to another member
func (h *UnionHandler) handleTransferLeadership(client *Client, message *valueobject.Message) *valueobject.Response {
	userID := client.GetUserID()
	if userID == 0 {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeUnauthorized, "User not authenticated")
	}

	var req dto.TransferLeadershipRequest
	if err := json.Unmarshal(message.Data, &req); err != nil {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInvalidRequest, "Invalid transfer leadership data")
	}

	// Set current leader ID from authenticated user
	req.CurrentLeaderID = userID

	err := h.unionService.TransferLeadership(&req)
	if err != nil {
		log.Printf("Transfer leadership error: %v", err)
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeValidationError, err.Error())
	}

	return valueobject.NewSuccessResponseWithUniqueID(message.Type, message.Action, map[string]interface{}{
		"message": "会长职位转让成功",
	})
}

// handleGetUnionMembers handles getting union member list
func (h *UnionHandler) handleGetUnionMembers(client *Client, message *valueobject.Message) *valueobject.Response {
	// 注意：此接口无权限限制，任何人都可以查看工会成员列表
	var req dto.GetUnionMembersRequest
	if err := json.Unmarshal(message.Data, &req); err != nil {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInvalidRequest, "Invalid get union members data")
	}

	members, err := h.unionService.GetUnionMembers(&req)
	if err != nil {
		log.Printf("Get union members error: %v", err)
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeValidationError, err.Error())
	}

	return valueobject.NewSuccessResponseWithUniqueID(message.Type, message.Action, members)
}