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
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInvalidRequest, "Unknown auth action")
	}
}

func (h *AuthHandler) handleLogin(client *Client, message *valueobject.Message) *valueobject.Response {
	var req dto.LoginRequest
	if err := json.Unmarshal(message.Data, &req); err != nil {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInvalidRequest, "Invalid login data")
	}

	response, err := h.authService.Login(&req)
	if err != nil {
		// Check for specific error types to return appropriate error codes
		errorMsg := err.Error()
		if errorMsg == "user is already logged in" {
			return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeConflict, err.Error())
		} else if errorMsg == "invalid username or password" {
			return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeUnauthorized, err.Error())
		}
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInternalError, err.Error())
	}

	// Set client authentication
	client.SetAuth(true)
	client.SetUserID(response.UserID)
	client.Hub.SetUserClient(response.UserID, client)

	return valueobject.NewSuccessResponseWithUniqueID(message.Type, message.Action, response)
}

func (h *AuthHandler) handleRegister(client *Client, message *valueobject.Message) *valueobject.Response {
	var req dto.RegisterRequest
	if err := json.Unmarshal(message.Data, &req); err != nil {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInvalidRequest, "Invalid register data")
	}

	response, err := h.authService.Register(&req)
	if err != nil {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeValidationError, err.Error())
	}

	return valueobject.NewSuccessResponseWithUniqueID(message.Type, message.Action, response)
}

func (h *AuthHandler) handleLogout(client *Client, message *valueobject.Message) *valueobject.Response {
	if err := h.authService.Logout(client.GetUserID()); err != nil {
		log.Printf("Logout error: %v", err)
	}

	// Clear client authentication
	client.Hub.RemoveUserClient(client.GetUserID())
	client.SetAuth(false)
	client.SetUserID(0)

	return valueobject.NewSuccessResponseWithUniqueID(message.Type, message.Action, map[string]string{"message": "Logged out successfully"})
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
		return valueobject.NewSuccessResponseWithUniqueID(message.Type, message.Action, map[string]string{"pong": "pong"})
	}
	return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInvalidRequest, "Unknown heartbeat action")
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
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInvalidRequest, "Unknown player action")
	}
}

func (h *PlayerHandler) handleGetPlayerInfo(client *Client, message *valueobject.Message) *valueobject.Response {
	response, err := h.playerService.GetPlayerInfo(client.GetUserID())
	if err != nil {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInternalError, err.Error())
	}
	return valueobject.NewSuccessResponseWithUniqueID(message.Type, message.Action, response)
}

func (h *PlayerHandler) handleUpdatePlayer(client *Client, message *valueobject.Message) *valueobject.Response {
	var req dto.UpdatePlayerRequest
	if err := json.Unmarshal(message.Data, &req); err != nil {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInvalidRequest, "Invalid update data")
	}

	req.UserID = client.GetUserID() // Ensure user can only update their own data
	if err := h.playerService.UpdatePlayer(&req); err != nil {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInternalError, err.Error())
	}

	return valueobject.NewSuccessResponseWithUniqueID(message.Type, message.Action, map[string]string{"message": "Player updated successfully"})
}

func (h *PlayerHandler) handleGetEquipment(client *Client, message *valueobject.Message) *valueobject.Response {
	equipment, err := h.playerService.GetUserEquipment(client.GetUserID())
	if err != nil {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInternalError, err.Error())
	}
	return valueobject.NewSuccessResponseWithUniqueID(message.Type, message.Action, equipment)
}

func (h *PlayerHandler) handleSaveEquipment(client *Client, message *valueobject.Message) *valueobject.Response {
	var req dto.SaveEquipmentRequest
	if err := json.Unmarshal(message.Data, &req); err != nil {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInvalidRequest, "Invalid equipment data")
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
			return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeValidationError, err.Error())
		}
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInternalError, err.Error())
	}

	return valueobject.NewSuccessResponseWithUniqueID(message.Type, message.Action, equipment)
}

func (h *PlayerHandler) handleDeleteEquipment(client *Client, message *valueobject.Message) *valueobject.Response {
	var req struct {
		EquipID int `json:"equipid"`
	}
	if err := json.Unmarshal(message.Data, &req); err != nil {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInvalidRequest, "Invalid delete data")
	}

	if err := h.playerService.DeleteEquipment(req.EquipID, client.GetUserID()); err != nil {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInternalError, err.Error())
	}

	return valueobject.NewSuccessResponseWithUniqueID(message.Type, message.Action, map[string]string{"message": "Equipment deleted successfully"})
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
	case valueobject.ActionFriendRequest:
		return h.handleFriendRequest(client, message)
	case valueobject.ActionFriendResponse:
		return h.handleFriendResponse(client, message)
	default:
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInvalidRequest, "Unknown friend action")
	}
}

func (h *FriendHandler) handleGetFriends(client *Client, message *valueobject.Message) *valueobject.Response {
	friends, err := h.friendService.GetFriends(client.GetUserID())
	if err != nil {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInternalError, err.Error())
	}
	return valueobject.NewSuccessResponseWithUniqueID(message.Type, message.Action, friends)
}

func (h *FriendHandler) handleAddFriend(client *Client, message *valueobject.Message) *valueobject.Response {
	var req dto.AddFriendRequest
	if err := json.Unmarshal(message.Data, &req); err != nil {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInvalidRequest, "Invalid friend request data")
	}

	if err := h.friendService.SendFriendRequest(client.GetUserID(), &req); err != nil {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInternalError, err.Error())
	}

	return valueobject.NewSuccessResponseWithUniqueID(message.Type, message.Action, map[string]string{"message": "Friend request sent"})
}

func (h *FriendHandler) handleRemoveFriend(client *Client, message *valueobject.Message) *valueobject.Response {
	var req dto.RemoveFriendRequest
	if err := json.Unmarshal(message.Data, &req); err != nil {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInvalidRequest, "Invalid remove friend data")
	}

	if err := h.friendService.RemoveFriend(client.GetUserID(), &req); err != nil {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInternalError, err.Error())
	}

	return valueobject.NewSuccessResponseWithUniqueID(message.Type, message.Action, map[string]string{"message": "Friend removed"})
}

func (h *FriendHandler) handleAcceptFriend(client *Client, message *valueobject.Message) *valueobject.Response {
	var req dto.FriendActionRequest
	if err := json.Unmarshal(message.Data, &req); err != nil {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInvalidRequest, "Invalid accept friend data")
	}

	if err := h.friendService.AcceptFriendRequest(client.GetUserID(), &req); err != nil {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInternalError, err.Error())
	}

	return valueobject.NewSuccessResponseWithUniqueID(message.Type, message.Action, map[string]string{"message": "Friend request accepted"})
}

func (h *FriendHandler) handleRejectFriend(client *Client, message *valueobject.Message) *valueobject.Response {
	var req dto.FriendActionRequest
	if err := json.Unmarshal(message.Data, &req); err != nil {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInvalidRequest, "Invalid reject friend data")
	}

	if err := h.friendService.RejectFriendRequest(client.GetUserID(), &req); err != nil {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInternalError, err.Error())
	}

	return valueobject.NewSuccessResponseWithUniqueID(message.Type, message.Action, map[string]string{"message": "Friend request rejected"})
}

func (h *FriendHandler) handleGetFriendRank(client *Client, message *valueobject.Message) *valueobject.Response {
	ranking, err := h.friendService.GetFriendRanking(client.GetUserID())
	if err != nil {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInternalError, err.Error())
	}
	return valueobject.NewSuccessResponseWithUniqueID(message.Type, message.Action, ranking)
}

func (h *FriendHandler) handleFriendRequest(client *Client, message *valueobject.Message) *valueobject.Response {
	requests, err := h.friendService.GetFriendRequests(client.GetUserID())
	if err != nil {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInternalError, err.Error())
	}
	return valueobject.NewSuccessResponseWithUniqueID(message.Type, message.Action, requests)
}

func (h *FriendHandler) handleFriendResponse(client *Client, message *valueobject.Message) *valueobject.Response {
	var req dto.FriendResponseRequest
	if err := json.Unmarshal(message.Data, &req); err != nil {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInvalidRequest, "Invalid friend response data")
	}

	if req.FromUserID <= 0 {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInvalidRequest, "FromUserID must be positive")
	}

	if req.Accept {
		err := h.friendService.AcceptFriendRequestByUserID(client.GetUserID(), req.FromUserID)
		if err != nil {
			return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInternalError, err.Error())
		}
		return valueobject.NewSuccessResponseWithUniqueID(message.Type, message.Action, map[string]string{"message": "Friend request accepted"})
	} else {
		err := h.friendService.RejectFriendRequestByUserID(client.GetUserID(), req.FromUserID)
		if err != nil {
			return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInternalError, err.Error())
		}
		return valueobject.NewSuccessResponseWithUniqueID(message.Type, message.Action, map[string]string{"message": "Friend request rejected"})
	}
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
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInvalidRequest, "Unknown ranking action")
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
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInternalError, err.Error())
	}
	return valueobject.NewSuccessResponseWithUniqueID(message.Type, message.Action, ranking)
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
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInternalError, err.Error())
	}
	return valueobject.NewSuccessResponseWithUniqueID(message.Type, message.Action, ranking)
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
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInvalidRequest, "Unknown user equipment action")
	}
}

func (h *UserEquipHandler) handleGetEquippedItems(client *Client, message *valueobject.Message) *valueobject.Response {
	equippedItems, err := h.userEquipService.GetUserEquippedItems(client.GetUserID())
	if err != nil {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInternalError, err.Error())
	}
	return valueobject.NewSuccessResponseWithUniqueID(message.Type, message.Action, equippedItems)
}

func (h *UserEquipHandler) handleEquipItem(client *Client, message *valueobject.Message) *valueobject.Response {
	var req struct {
		EquipSlot string `json:"equip_slot"`
		EquipID   int    `json:"equipid"`
	}
	if err := json.Unmarshal(message.Data, &req); err != nil {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInvalidRequest, "Invalid equip item data")
	}

	if req.EquipSlot == "" || req.EquipID <= 0 {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInvalidRequest, "Equipment slot and equipment ID are required")
	}

	err := h.userEquipService.EquipItem(client.GetUserID(), req.EquipSlot, req.EquipID)
	if err != nil {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInternalError, err.Error())
	}

	return valueobject.NewSuccessResponseWithUniqueID(message.Type, message.Action, map[string]string{"message": "Item equipped successfully"})
}

func (h *UserEquipHandler) handleUnequipItem(client *Client, message *valueobject.Message) *valueobject.Response {
	var req struct {
		EquipSlot string `json:"equip_slot"`
	}
	if err := json.Unmarshal(message.Data, &req); err != nil {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInvalidRequest, "Invalid unequip item data")
	}

	if req.EquipSlot == "" {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInvalidRequest, "Equipment slot is required")
	}

	err := h.userEquipService.UnequipItem(client.GetUserID(), req.EquipSlot)
	if err != nil {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInternalError, err.Error())
	}

	return valueobject.NewSuccessResponseWithUniqueID(message.Type, message.Action, map[string]string{"message": "Item unequipped successfully"})
}

func (h *UserEquipHandler) handleGetEquipmentStats(client *Client, message *valueobject.Message) *valueobject.Response {
	stats, err := h.userEquipService.GetEquipmentStats(client.GetUserID())
	if err != nil {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInternalError, err.Error())
	}
	return valueobject.NewSuccessResponseWithUniqueID(message.Type, message.Action, stats)
}

func (h *UserEquipHandler) handleGetEquippedBySlot(client *Client, message *valueobject.Message) *valueobject.Response {
	var req struct {
		EquipSlot string `json:"equip_slot"`
	}
	if err := json.Unmarshal(message.Data, &req); err != nil {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInvalidRequest, "Invalid slot data")
	}

	if req.EquipSlot == "" {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInvalidRequest, "Equipment slot is required")
	}

	equipment, err := h.userEquipService.GetEquippedItemsBySlot(client.GetUserID(), req.EquipSlot)
	if err != nil {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInternalError, err.Error())
	}
	return valueobject.NewSuccessResponseWithUniqueID(message.Type, message.Action, equipment)
}

// ExperienceHandler handles experience-related messages
type ExperienceHandler struct {
	experienceService ExperienceServiceInterface
}

// NewExperienceHandler creates a new experience handler
func NewExperienceHandler(experienceService ExperienceServiceInterface) *ExperienceHandler {
	return &ExperienceHandler{experienceService: experienceService}
}

// Handle handles experience messages
func (h *ExperienceHandler) Handle(client *Client, message *valueobject.Message) *valueobject.Response {
	switch message.Action {
	case valueobject.ActionGetByLevel:
		return h.handleGetByLevel(client, message)
	case valueobject.ActionGetAllLevels:
		return h.handleGetAllLevels(client, message)
	default:
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInvalidRequest, "Unknown experience action")
	}
}

func (h *ExperienceHandler) handleGetByLevel(client *Client, message *valueobject.Message) *valueobject.Response {
	var req struct {
		Level int `json:"level"`
	}
	if err := json.Unmarshal(message.Data, &req); err != nil {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInvalidRequest, "Invalid level data")
	}

	if req.Level <= 0 {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInvalidRequest, "Level must be positive")
	}

	experience, err := h.experienceService.GetByLevel(req.Level)
	if err != nil {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInternalError, err.Error())
	}
	if experience == nil {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeNotFound, "Experience level not found")
	}
	return valueobject.NewSuccessResponseWithUniqueID(message.Type, message.Action, experience)
}

func (h *ExperienceHandler) handleGetAllLevels(client *Client, message *valueobject.Message) *valueobject.Response {
	experiences, err := h.experienceService.GetAllLevels()
	if err != nil {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInternalError, err.Error())
	}
	return valueobject.NewSuccessResponseWithUniqueID(message.Type, message.Action, experiences)
}

// WeaponHandler handles weapon-related messages
type WeaponHandler struct {
	weaponService WeaponServiceInterface
}

// NewWeaponHandler creates a new weapon handler
func NewWeaponHandler(weaponService WeaponServiceInterface) *WeaponHandler {
	return &WeaponHandler{weaponService: weaponService}
}

// Handle handles weapon messages
func (h *WeaponHandler) Handle(client *Client, message *valueobject.Message) *valueobject.Response {
	switch message.Action {
	case valueobject.ActionGetWeapon:
		return h.handleGetWeapon(client, message)
	case valueobject.ActionGetAllWeapons:
		return h.handleGetAllWeapons(client, message)
	case valueobject.ActionCreateWeapon:
		return h.handleCreateWeapon(client, message)
	case valueobject.ActionUpdateWeapon:
		return h.handleUpdateWeapon(client, message)
	case valueobject.ActionDeleteWeapon:
		return h.handleDeleteWeapon(client, message)
	default:
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInvalidRequest, "Unknown weapon action")
	}
}

func (h *WeaponHandler) handleGetWeapon(client *Client, message *valueobject.Message) *valueobject.Response {
	var req dto.GetWeaponRequest
	if err := json.Unmarshal(message.Data, &req); err != nil {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInvalidRequest, "Invalid weapon ID data")
	}

	if req.WeaponID <= 0 {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInvalidRequest, "Weapon ID must be positive")
	}

	weapon, err := h.weaponService.GetWeaponByID(req.WeaponID)
	if err != nil {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInternalError, err.Error())
	}
	return valueobject.NewSuccessResponseWithUniqueID(message.Type, message.Action, weapon)
}

func (h *WeaponHandler) handleGetAllWeapons(client *Client, message *valueobject.Message) *valueobject.Response {
	weapons, err := h.weaponService.GetAllWeapons()
	if err != nil {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInternalError, err.Error())
	}
	return valueobject.NewSuccessResponseWithUniqueID(message.Type, message.Action, weapons)
}

func (h *WeaponHandler) handleCreateWeapon(client *Client, message *valueobject.Message) *valueobject.Response {
	var req dto.CreateWeaponRequest
	if err := json.Unmarshal(message.Data, &req); err != nil {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInvalidRequest, "Invalid create weapon data")
	}

	weapon, err := h.weaponService.CreateWeapon(&req)
	if err != nil {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeValidationError, err.Error())
	}
	return valueobject.NewSuccessResponseWithUniqueID(message.Type, message.Action, weapon)
}

func (h *WeaponHandler) handleUpdateWeapon(client *Client, message *valueobject.Message) *valueobject.Response {
	var req dto.UpdateWeaponRequest
	if err := json.Unmarshal(message.Data, &req); err != nil {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInvalidRequest, "Invalid update weapon data")
	}

	weapon, err := h.weaponService.UpdateWeapon(&req)
	if err != nil {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeValidationError, err.Error())
	}
	return valueobject.NewSuccessResponseWithUniqueID(message.Type, message.Action, weapon)
}

func (h *WeaponHandler) handleDeleteWeapon(client *Client, message *valueobject.Message) *valueobject.Response {
	var req dto.DeleteWeaponRequest
	if err := json.Unmarshal(message.Data, &req); err != nil {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInvalidRequest, "Invalid delete weapon data")
	}

	if req.WeaponID <= 0 {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInvalidRequest, "Weapon ID must be positive")
	}

	err := h.weaponService.DeleteWeapon(req.WeaponID)
	if err != nil {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeValidationError, err.Error())
	}
	return valueobject.NewSuccessResponseWithUniqueID(message.Type, message.Action, map[string]string{"message": "Weapon deleted successfully"})
}

// UserWeaponHandler handles user weapon messages
type UserWeaponHandler struct {
	userWeaponService UserWeaponServiceInterface
}

// NewUserWeaponHandler creates a new user weapon handler
func NewUserWeaponHandler(userWeaponService UserWeaponServiceInterface) *UserWeaponHandler {
	return &UserWeaponHandler{userWeaponService: userWeaponService}
}

// Handle handles user weapon messages
func (h *UserWeaponHandler) Handle(client *Client, message *valueobject.Message) *valueobject.Response {
	if !client.IsAuthenticated() {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeUnauthorized, "Authentication required")
	}

	switch message.Action {
	case valueobject.ActionGetUserWeapons:
		return h.handleGetUserWeapons(client, message)
	case valueobject.ActionAddUserWeapon:
		return h.handleAddUserWeapon(client, message)
	case valueobject.ActionRemoveUserWeapon:
		return h.handleRemoveUserWeapon(client, message)
	case valueobject.ActionCheckUserWeapon:
		return h.handleCheckUserWeapon(client, message)
	default:
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInvalidRequest, "Unknown user weapon action")
	}
}

func (h *UserWeaponHandler) handleGetUserWeapons(client *Client, message *valueobject.Message) *valueobject.Response {
	var req dto.GetUserWeaponsRequest
	if err := json.Unmarshal(message.Data, &req); err != nil {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInvalidRequest, "Invalid get user weapons data")
	}

	// If no user ID provided, use client's user ID
	userID := req.UserID
	if userID <= 0 {
		userID = client.GetUserID()
	}

	weapons, err := h.userWeaponService.GetUserWeapons(userID, req.WithDetails)
	if err != nil {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInternalError, err.Error())
	}
	return valueobject.NewSuccessResponseWithUniqueID(message.Type, message.Action, weapons)
}

func (h *UserWeaponHandler) handleAddUserWeapon(client *Client, message *valueobject.Message) *valueobject.Response {
	var req dto.AddUserWeaponRequest
	if err := json.Unmarshal(message.Data, &req); err != nil {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInvalidRequest, "Invalid add user weapon data")
	}

	// If no user ID provided, use client's user ID
	if req.UserID <= 0 {
		req.UserID = client.GetUserID()
	}

	userWeapon, err := h.userWeaponService.AddUserWeapon(&req)
	if err != nil {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeValidationError, err.Error())
	}
	return valueobject.NewSuccessResponseWithUniqueID(message.Type, message.Action, userWeapon)
}

func (h *UserWeaponHandler) handleRemoveUserWeapon(client *Client, message *valueobject.Message) *valueobject.Response {
	var req dto.RemoveUserWeaponRequest
	if err := json.Unmarshal(message.Data, &req); err != nil {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInvalidRequest, "Invalid remove user weapon data")
	}

	// If no user ID provided, use client's user ID
	if req.UserID <= 0 {
		req.UserID = client.GetUserID()
	}

	err := h.userWeaponService.RemoveUserWeapon(&req)
	if err != nil {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeValidationError, err.Error())
	}
	return valueobject.NewSuccessResponseWithUniqueID(message.Type, message.Action, map[string]string{"message": "User weapon removed successfully"})
}

func (h *UserWeaponHandler) handleCheckUserWeapon(client *Client, message *valueobject.Message) *valueobject.Response {
	var req dto.CheckUserWeaponRequest
	if err := json.Unmarshal(message.Data, &req); err != nil {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInvalidRequest, "Invalid check user weapon data")
	}

	// If no user ID provided, use client's user ID
	if req.UserID <= 0 {
		req.UserID = client.GetUserID()
	}

	result, err := h.userWeaponService.CheckUserWeapon(&req)
	if err != nil {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInternalError, err.Error())
	}
	return valueobject.NewSuccessResponseWithUniqueID(message.Type, message.Action, result)
}