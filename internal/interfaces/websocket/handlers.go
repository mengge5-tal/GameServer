package websocket

import (
	"GameServer/internal/application/dto"
	"GameServer/internal/domain/valueobject"
	"encoding/json"
	"log"
)

// AuthHandler handles authentication messages
type AuthHandler struct {
	authService AuthServiceInterface
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService AuthServiceInterface) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Handle handles authentication messages
func (h *AuthHandler) Handle(client *Client, message *valueobject.Message) *valueobject.Response {
	switch message.Action {
	case valueobject.ActionLogin:
		return h.handleLogin(client, message)
	case valueobject.ActionRegister:
		return h.handleRegister(client, message)
	case valueobject.ActionLogout:
		return h.handleLogout(client, message)
	default:
		return valueobject.NewErrorResponse(message.RequestID, valueobject.CodeInvalidRequest, "Unknown auth action")
	}
}

func (h *AuthHandler) handleLogin(client *Client, message *valueobject.Message) *valueobject.Response {
	var req dto.LoginRequest
	if err := json.Unmarshal(message.Data, &req); err != nil {
		return valueobject.NewErrorResponse(message.RequestID, valueobject.CodeInvalidRequest, "Invalid login data")
	}

	response, err := h.authService.Login(&req)
	if err != nil {
		// Check for specific error types to return appropriate error codes
		errorMsg := err.Error()
		if errorMsg == "user is already logged in" {
			return valueobject.NewErrorResponse(message.RequestID, valueobject.CodeConflict, err.Error())
		} else if errorMsg == "invalid username or password" {
			return valueobject.NewErrorResponse(message.RequestID, valueobject.CodeUnauthorized, err.Error())
		}
		return valueobject.NewErrorResponse(message.RequestID, valueobject.CodeInternalError, err.Error())
	}

	// Set client authentication
	client.SetAuth(true)
	client.SetUserID(response.UserID)
	client.Hub.SetUserClient(response.UserID, client)

	return valueobject.NewSuccessResponse(message.RequestID, response)
}

func (h *AuthHandler) handleRegister(client *Client, message *valueobject.Message) *valueobject.Response {
	var req dto.RegisterRequest
	if err := json.Unmarshal(message.Data, &req); err != nil {
		return valueobject.NewErrorResponse(message.RequestID, valueobject.CodeInvalidRequest, "Invalid register data")
	}

	response, err := h.authService.Register(&req)
	if err != nil {
		return valueobject.NewErrorResponse(message.RequestID, valueobject.CodeValidationError, err.Error())
	}

	return valueobject.NewSuccessResponse(message.RequestID, response)
}

func (h *AuthHandler) handleLogout(client *Client, message *valueobject.Message) *valueobject.Response {
	if err := h.authService.Logout(client.GetUserID()); err != nil {
		log.Printf("Logout error: %v", err)
	}

	// Clear client authentication
	client.Hub.RemoveUserClient(client.GetUserID())
	client.SetAuth(false)
	client.SetUserID(0)

	return valueobject.NewSuccessResponse(message.RequestID, map[string]string{"message": "Logged out successfully"})
}

// HeartbeatHandler handles heartbeat messages
type HeartbeatHandler struct{}

// NewHeartbeatHandler creates a new heartbeat handler
func NewHeartbeatHandler() *HeartbeatHandler {
	return &HeartbeatHandler{}
}

// Handle handles heartbeat messages
func (h *HeartbeatHandler) Handle(client *Client, message *valueobject.Message) *valueobject.Response {
	if message.Action == valueobject.ActionPing {
		return valueobject.NewSuccessResponse(message.RequestID, map[string]string{"pong": "pong"})
	}
	return valueobject.NewErrorResponse(message.RequestID, valueobject.CodeInvalidRequest, "Unknown heartbeat action")
}

// PlayerHandler handles player-related messages
type PlayerHandler struct {
	playerService PlayerServiceInterface
}

// NewPlayerHandler creates a new player handler
func NewPlayerHandler(playerService PlayerServiceInterface) *PlayerHandler {
	return &PlayerHandler{playerService: playerService}
}

// Handle handles player messages
func (h *PlayerHandler) Handle(client *Client, message *valueobject.Message) *valueobject.Response {
	switch message.Action {
	case valueobject.ActionGetPlayerInfo:
		return h.handleGetPlayerInfo(client, message)
	case valueobject.ActionUpdatePlayer:
		return h.handleUpdatePlayer(client, message)
	case valueobject.ActionGetEquip:
		return h.handleGetEquipment(client, message)
	case valueobject.ActionSaveEquip:
		return h.handleSaveEquipment(client, message)
	case valueobject.ActionDeleteEquip:
		return h.handleDeleteEquipment(client, message)
	case valueobject.ActionDelEquip:
		return h.handleDeleteEquipment(client, message)
	default:
		return valueobject.NewErrorResponse(message.RequestID, valueobject.CodeInvalidRequest, "Unknown player action")
	}
}

func (h *PlayerHandler) handleGetPlayerInfo(client *Client, message *valueobject.Message) *valueobject.Response {
	response, err := h.playerService.GetPlayerInfo(client.GetUserID())
	if err != nil {
		return valueobject.NewErrorResponse(message.RequestID, valueobject.CodeInternalError, err.Error())
	}
	return valueobject.NewSuccessResponse(message.RequestID, response)
}

func (h *PlayerHandler) handleUpdatePlayer(client *Client, message *valueobject.Message) *valueobject.Response {
	var req dto.UpdatePlayerRequest
	if err := json.Unmarshal(message.Data, &req); err != nil {
		return valueobject.NewErrorResponse(message.RequestID, valueobject.CodeInvalidRequest, "Invalid update data")
	}

	req.UserID = client.GetUserID() // Ensure user can only update their own data
	if err := h.playerService.UpdatePlayer(&req); err != nil {
		return valueobject.NewErrorResponse(message.RequestID, valueobject.CodeInternalError, err.Error())
	}

	return valueobject.NewSuccessResponse(message.RequestID, map[string]string{"message": "Player updated successfully"})
}

func (h *PlayerHandler) handleGetEquipment(client *Client, message *valueobject.Message) *valueobject.Response {
	equipment, err := h.playerService.GetUserEquipment(client.GetUserID())
	if err != nil {
		return valueobject.NewErrorResponse(message.RequestID, valueobject.CodeInternalError, err.Error())
	}
	return valueobject.NewSuccessResponse(message.RequestID, equipment)
}

func (h *PlayerHandler) handleSaveEquipment(client *Client, message *valueobject.Message) *valueobject.Response {
	var req dto.SaveEquipmentRequest
	if err := json.Unmarshal(message.Data, &req); err != nil {
		return valueobject.NewErrorResponse(message.RequestID, valueobject.CodeInvalidRequest, "Invalid equipment data")
	}

	req.UserID = client.GetUserID() // Ensure user can only save their own equipment
	equipment, err := h.playerService.SaveEquipment(&req)
	if err != nil {
		// Check for specific error types
		errorMsg := err.Error()
		if errorMsg == "type and quality must be positive integers" || 
		   errorMsg == "equipment not found for update" ||
		   errorMsg == "unauthorized to update this equipment" ||
		   errorMsg == "equipment sequence limit reached for this type and quality" {
			return valueobject.NewErrorResponse(message.RequestID, valueobject.CodeValidationError, err.Error())
		}
		return valueobject.NewErrorResponse(message.RequestID, valueobject.CodeInternalError, err.Error())
	}

	return valueobject.NewSuccessResponse(message.RequestID, equipment)
}

func (h *PlayerHandler) handleDeleteEquipment(client *Client, message *valueobject.Message) *valueobject.Response {
	var req struct {
		EquipID int `json:"equipid"`
	}
	if err := json.Unmarshal(message.Data, &req); err != nil {
		return valueobject.NewErrorResponse(message.RequestID, valueobject.CodeInvalidRequest, "Invalid delete data")
	}

	if err := h.playerService.DeleteEquipment(req.EquipID, client.GetUserID()); err != nil {
		return valueobject.NewErrorResponse(message.RequestID, valueobject.CodeInternalError, err.Error())
	}

	return valueobject.NewSuccessResponse(message.RequestID, map[string]string{"message": "Equipment deleted successfully"})
}

// FriendHandler handles friend-related messages
type FriendHandler struct {
	friendService FriendServiceInterface
}

// NewFriendHandler creates a new friend handler
func NewFriendHandler(friendService FriendServiceInterface) *FriendHandler {
	return &FriendHandler{friendService: friendService}
}

// Handle handles friend messages
func (h *FriendHandler) Handle(client *Client, message *valueobject.Message) *valueobject.Response {
	switch message.Action {
	case valueobject.ActionGetFriends:
		return h.handleGetFriends(client, message)
	case valueobject.ActionAddFriend:
		return h.handleAddFriend(client, message)
	case valueobject.ActionRemoveFriend:
		return h.handleRemoveFriend(client, message)
	case valueobject.ActionAcceptFriend:
		return h.handleAcceptFriend(client, message)
	case valueobject.ActionRejectFriend:
		return h.handleRejectFriend(client, message)
	case valueobject.ActionGetFriendRank:
		return h.handleGetFriendRank(client, message)
	default:
		return valueobject.NewErrorResponse(message.RequestID, valueobject.CodeInvalidRequest, "Unknown friend action")
	}
}

func (h *FriendHandler) handleGetFriends(client *Client, message *valueobject.Message) *valueobject.Response {
	friends, err := h.friendService.GetFriends(client.GetUserID())
	if err != nil {
		return valueobject.NewErrorResponse(message.RequestID, valueobject.CodeInternalError, err.Error())
	}
	return valueobject.NewSuccessResponse(message.RequestID, friends)
}

func (h *FriendHandler) handleAddFriend(client *Client, message *valueobject.Message) *valueobject.Response {
	var req dto.AddFriendRequest
	if err := json.Unmarshal(message.Data, &req); err != nil {
		return valueobject.NewErrorResponse(message.RequestID, valueobject.CodeInvalidRequest, "Invalid friend request data")
	}

	if err := h.friendService.SendFriendRequest(client.GetUserID(), &req); err != nil {
		return valueobject.NewErrorResponse(message.RequestID, valueobject.CodeInternalError, err.Error())
	}

	return valueobject.NewSuccessResponse(message.RequestID, map[string]string{"message": "Friend request sent"})
}

func (h *FriendHandler) handleRemoveFriend(client *Client, message *valueobject.Message) *valueobject.Response {
	var req dto.RemoveFriendRequest
	if err := json.Unmarshal(message.Data, &req); err != nil {
		return valueobject.NewErrorResponse(message.RequestID, valueobject.CodeInvalidRequest, "Invalid remove friend data")
	}

	if err := h.friendService.RemoveFriend(client.GetUserID(), &req); err != nil {
		return valueobject.NewErrorResponse(message.RequestID, valueobject.CodeInternalError, err.Error())
	}

	return valueobject.NewSuccessResponse(message.RequestID, map[string]string{"message": "Friend removed"})
}

func (h *FriendHandler) handleAcceptFriend(client *Client, message *valueobject.Message) *valueobject.Response {
	var req dto.FriendActionRequest
	if err := json.Unmarshal(message.Data, &req); err != nil {
		return valueobject.NewErrorResponse(message.RequestID, valueobject.CodeInvalidRequest, "Invalid accept friend data")
	}

	if err := h.friendService.AcceptFriendRequest(client.GetUserID(), &req); err != nil {
		return valueobject.NewErrorResponse(message.RequestID, valueobject.CodeInternalError, err.Error())
	}

	return valueobject.NewSuccessResponse(message.RequestID, map[string]string{"message": "Friend request accepted"})
}

func (h *FriendHandler) handleRejectFriend(client *Client, message *valueobject.Message) *valueobject.Response {
	var req dto.FriendActionRequest
	if err := json.Unmarshal(message.Data, &req); err != nil {
		return valueobject.NewErrorResponse(message.RequestID, valueobject.CodeInvalidRequest, "Invalid reject friend data")
	}

	if err := h.friendService.RejectFriendRequest(client.GetUserID(), &req); err != nil {
		return valueobject.NewErrorResponse(message.RequestID, valueobject.CodeInternalError, err.Error())
	}

	return valueobject.NewSuccessResponse(message.RequestID, map[string]string{"message": "Friend request rejected"})
}

func (h *FriendHandler) handleGetFriendRank(client *Client, message *valueobject.Message) *valueobject.Response {
	ranking, err := h.friendService.GetFriendRanking(client.GetUserID())
	if err != nil {
		return valueobject.NewErrorResponse(message.RequestID, valueobject.CodeInternalError, err.Error())
	}
	return valueobject.NewSuccessResponse(message.RequestID, ranking)
}

// RankingHandler handles ranking-related messages
type RankingHandler struct {
	rankingService RankingServiceInterface
}

// NewRankingHandler creates a new ranking handler
func NewRankingHandler(rankingService RankingServiceInterface) *RankingHandler {
	return &RankingHandler{rankingService: rankingService}
}

// Handle handles ranking messages
func (h *RankingHandler) Handle(client *Client, message *valueobject.Message) *valueobject.Response {
	switch message.Action {
	case valueobject.ActionGetAllRank:
		return h.handleGetAllRank(client, message)
	case valueobject.ActionGetRank:
		return h.handleGetRank(client, message)
	default:
		return valueobject.NewErrorResponse(message.RequestID, valueobject.CodeInvalidRequest, "Unknown ranking action")
	}
}

func (h *RankingHandler) handleGetAllRank(client *Client, message *valueobject.Message) *valueobject.Response {
	var req dto.GetRankingRequest
	if err := json.Unmarshal(message.Data, &req); err != nil {
		// Set default values if no data provided
		req.RankType = "level"
		req.Limit = 50
	}

	ranking, err := h.rankingService.GetRanking(&req)
	if err != nil {
		return valueobject.NewErrorResponse(message.RequestID, valueobject.CodeInternalError, err.Error())
	}
	return valueobject.NewSuccessResponse(message.RequestID, ranking)
}

func (h *RankingHandler) handleGetRank(client *Client, message *valueobject.Message) *valueobject.Response {
	var req struct {
		RankType string `json:"rank_type"`
	}
	if err := json.Unmarshal(message.Data, &req); err != nil {
		req.RankType = "level" // Default to level ranking
	}

	ranking, err := h.rankingService.GetUserRanking(client.GetUserID(), req.RankType)
	if err != nil {
		return valueobject.NewErrorResponse(message.RequestID, valueobject.CodeInternalError, err.Error())
	}
	return valueobject.NewSuccessResponse(message.RequestID, ranking)
}

// UserEquipHandler handles user equipment messages
type UserEquipHandler struct {
	userEquipService UserEquipServiceInterface
}

// NewUserEquipHandler creates a new user equipment handler
func NewUserEquipHandler(userEquipService UserEquipServiceInterface) *UserEquipHandler {
	return &UserEquipHandler{userEquipService: userEquipService}
}

// Handle handles user equipment messages
func (h *UserEquipHandler) Handle(client *Client, message *valueobject.Message) *valueobject.Response {
	switch message.Action {
	case valueobject.ActionGetEquippedItems:
		return h.handleGetEquippedItems(client, message)
	case valueobject.ActionEquipItem:
		return h.handleEquipItem(client, message)
	case valueobject.ActionUnequipItem:
		return h.handleUnequipItem(client, message)
	case valueobject.ActionGetEquipmentStats:
		return h.handleGetEquipmentStats(client, message)
	case valueobject.ActionGetEquippedBySlot:
		return h.handleGetEquippedBySlot(client, message)
	default:
		return valueobject.NewErrorResponse(message.RequestID, valueobject.CodeInvalidRequest, "Unknown user equipment action")
	}
}

func (h *UserEquipHandler) handleGetEquippedItems(client *Client, message *valueobject.Message) *valueobject.Response {
	equippedItems, err := h.userEquipService.GetUserEquippedItems(client.GetUserID())
	if err != nil {
		return valueobject.NewErrorResponse(message.RequestID, valueobject.CodeInternalError, err.Error())
	}
	return valueobject.NewSuccessResponse(message.RequestID, equippedItems)
}

func (h *UserEquipHandler) handleEquipItem(client *Client, message *valueobject.Message) *valueobject.Response {
	var req struct {
		EquipSlot string `json:"equip_slot"`
		EquipID   int    `json:"equipid"`
	}
	if err := json.Unmarshal(message.Data, &req); err != nil {
		return valueobject.NewErrorResponse(message.RequestID, valueobject.CodeInvalidRequest, "Invalid equip item data")
	}

	if req.EquipSlot == "" || req.EquipID <= 0 {
		return valueobject.NewErrorResponse(message.RequestID, valueobject.CodeInvalidRequest, "Equipment slot and equipment ID are required")
	}

	err := h.userEquipService.EquipItem(client.GetUserID(), req.EquipSlot, req.EquipID)
	if err != nil {
		return valueobject.NewErrorResponse(message.RequestID, valueobject.CodeInternalError, err.Error())
	}

	return valueobject.NewSuccessResponse(message.RequestID, map[string]string{"message": "Item equipped successfully"})
}

func (h *UserEquipHandler) handleUnequipItem(client *Client, message *valueobject.Message) *valueobject.Response {
	var req struct {
		EquipSlot string `json:"equip_slot"`
	}
	if err := json.Unmarshal(message.Data, &req); err != nil {
		return valueobject.NewErrorResponse(message.RequestID, valueobject.CodeInvalidRequest, "Invalid unequip item data")
	}

	if req.EquipSlot == "" {
		return valueobject.NewErrorResponse(message.RequestID, valueobject.CodeInvalidRequest, "Equipment slot is required")
	}

	err := h.userEquipService.UnequipItem(client.GetUserID(), req.EquipSlot)
	if err != nil {
		return valueobject.NewErrorResponse(message.RequestID, valueobject.CodeInternalError, err.Error())
	}

	return valueobject.NewSuccessResponse(message.RequestID, map[string]string{"message": "Item unequipped successfully"})
}

func (h *UserEquipHandler) handleGetEquipmentStats(client *Client, message *valueobject.Message) *valueobject.Response {
	stats, err := h.userEquipService.GetEquipmentStats(client.GetUserID())
	if err != nil {
		return valueobject.NewErrorResponse(message.RequestID, valueobject.CodeInternalError, err.Error())
	}
	return valueobject.NewSuccessResponse(message.RequestID, stats)
}

func (h *UserEquipHandler) handleGetEquippedBySlot(client *Client, message *valueobject.Message) *valueobject.Response {
	var req struct {
		EquipSlot string `json:"equip_slot"`
	}
	if err := json.Unmarshal(message.Data, &req); err != nil {
		return valueobject.NewErrorResponse(message.RequestID, valueobject.CodeInvalidRequest, "Invalid slot data")
	}

	if req.EquipSlot == "" {
		return valueobject.NewErrorResponse(message.RequestID, valueobject.CodeInvalidRequest, "Equipment slot is required")
	}

	equipment, err := h.userEquipService.GetEquippedItemsBySlot(client.GetUserID(), req.EquipSlot)
	if err != nil {
		return valueobject.NewErrorResponse(message.RequestID, valueobject.CodeInternalError, err.Error())
	}
	return valueobject.NewSuccessResponse(message.RequestID, equipment)
}