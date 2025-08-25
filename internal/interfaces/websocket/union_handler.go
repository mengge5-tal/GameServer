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