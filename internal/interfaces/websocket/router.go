package websocket

import (
	"GameServer/internal/domain/valueobject"
	"log"
)

// MessageRouter defines the interface for message routing
type MessageRouter interface {
	Handle(client *Client, message *valueobject.Message) *valueobject.Response
}

// messageRouter implements MessageRouter
type messageRouter struct {
	handlers map[valueobject.MessageType]map[valueobject.MessageAction]MessageHandler
	services *ServiceContainer
}

// MessageHandler defines the interface for message handlers
type MessageHandler interface {
	Handle(client *Client, message *valueobject.Message) *valueobject.Response
}

// MessageHandlerFunc is a function type that implements MessageHandler
type MessageHandlerFunc func(client *Client, message *valueobject.Message) *valueobject.Response

// Handle implements MessageHandler interface
func (f MessageHandlerFunc) Handle(client *Client, message *valueobject.Message) *valueobject.Response {
	return f(client, message)
}

// NewMessageRouter creates a new message router
func NewMessageRouter(services *ServiceContainer) MessageRouter {
	router := &messageRouter{
		handlers: make(map[valueobject.MessageType]map[valueobject.MessageAction]MessageHandler),
		services: services,
	}

	// Register handlers
	router.registerHandlers()

	return router
}

// Handle routes a message to the appropriate handler
func (r *messageRouter) Handle(client *Client, message *valueobject.Message) *valueobject.Response {
	// Find handler for message type
	typeHandlers, exists := r.handlers[message.Type]
	if !exists {
		log.Printf("Unknown message type: %s", message.Type)
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInvalidRequest, "Unknown message type")
	}

	// Find handler for action
	handler, exists := typeHandlers[message.Action]
	if !exists {
		log.Printf("Unknown action %s for type %s", message.Action, message.Type)
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeInvalidRequest, "Unknown action")
	}

	// Check authentication for protected actions
	if r.requiresAuth(message.Type, message.Action) && !client.IsAuthenticated() {
		return valueobject.NewErrorResponseWithUniqueID(message.Type, message.Action, valueobject.CodeUnauthorized, "Authentication required")
	}

	// Handle the message
	return handler.Handle(client, message)
}

// registerHandlers registers all message handlers
func (r *messageRouter) registerHandlers() {
	// Auth handlers
	r.register(valueobject.MessageTypeAuth, valueobject.ActionLogin, NewAuthHandler(r.services.AuthService))
	r.register(valueobject.MessageTypeAuth, valueobject.ActionRegister, NewAuthHandler(r.services.AuthService))
	r.register(valueobject.MessageTypeAuth, valueobject.ActionLogout, NewAuthHandler(r.services.AuthService))

	// Heartbeat handlers
	r.register(valueobject.MessageTypeHeartbeat, valueobject.ActionPing, NewHeartbeatHandler())

	// Player handlers
	r.register(valueobject.MessageTypePlayer, valueobject.ActionGetPlayerInfo, NewPlayerHandler(r.services.PlayerService))
	r.register(valueobject.MessageTypePlayer, valueobject.ActionUpdatePlayer, NewPlayerHandler(r.services.PlayerService))

	// Equipment handlers
	r.register(valueobject.MessageTypeEquip, valueobject.ActionGetEquip, NewPlayerHandler(r.services.PlayerService))
	r.register(valueobject.MessageTypeEquip, valueobject.ActionSaveEquip, NewPlayerHandler(r.services.PlayerService))
	r.register(valueobject.MessageTypeEquip, valueobject.ActionDeleteEquip, NewPlayerHandler(r.services.PlayerService))
	r.register(valueobject.MessageTypeEquip, valueobject.ActionDelEquip, NewPlayerHandler(r.services.PlayerService))
	r.register(valueobject.MessageTypeEquip, valueobject.ActionBatchDeleteEquip, NewPlayerHandler(r.services.PlayerService))

	// Friend handlers
	r.register(valueobject.MessageTypeFriend, valueobject.ActionGetFriends, NewFriendHandler(r.services.FriendService))
	r.register(valueobject.MessageTypeFriend, valueobject.ActionAddFriend, NewFriendHandler(r.services.FriendService))
	r.register(valueobject.MessageTypeFriend, valueobject.ActionRemoveFriend, NewFriendHandler(r.services.FriendService))
	r.register(valueobject.MessageTypeFriend, valueobject.ActionAcceptFriend, NewFriendHandler(r.services.FriendService))
	r.register(valueobject.MessageTypeFriend, valueobject.ActionRejectFriend, NewFriendHandler(r.services.FriendService))
	r.register(valueobject.MessageTypeFriend, valueobject.ActionGetFriendRank, NewFriendHandler(r.services.FriendService))
	r.register(valueobject.MessageTypeFriend, valueobject.ActionFriendRequest, NewFriendHandler(r.services.FriendService))
	r.register(valueobject.MessageTypeFriend, valueobject.ActionFriendResponse, NewFriendHandler(r.services.FriendService))
	r.register(valueobject.MessageTypeFriend, valueobject.ActionGetRecommendedFriends, NewFriendHandler(r.services.FriendService))


	// User Equipment handlers
	r.register(valueobject.MessageTypeUserEquip, valueobject.ActionGetEquippedItems, NewUserEquipHandler(r.services.UserEquipService))
	r.register(valueobject.MessageTypeUserEquip, valueobject.ActionEquipItem, NewUserEquipHandler(r.services.UserEquipService))
	r.register(valueobject.MessageTypeUserEquip, valueobject.ActionUnequipItem, NewUserEquipHandler(r.services.UserEquipService))
	r.register(valueobject.MessageTypeUserEquip, valueobject.ActionGetEquipmentStats, NewUserEquipHandler(r.services.UserEquipService))
	r.register(valueobject.MessageTypeUserEquip, valueobject.ActionGetEquippedBySlot, NewUserEquipHandler(r.services.UserEquipService))

	// Experience handlers
	r.register(valueobject.MessageTypeExperience, valueobject.ActionGetByLevel, NewExperienceHandler(r.services.ExperienceService))
	r.register(valueobject.MessageTypeExperience, valueobject.ActionGetAllLevels, NewExperienceHandler(r.services.ExperienceService))

	// Weapon handlers
	r.register(valueobject.MessageTypeWeapon, valueobject.ActionGetWeapon, NewWeaponHandler(r.services.WeaponService))
	r.register(valueobject.MessageTypeWeapon, valueobject.ActionGetAllWeapons, NewWeaponHandler(r.services.WeaponService))
	r.register(valueobject.MessageTypeWeapon, valueobject.ActionCreateWeapon, NewWeaponHandler(r.services.WeaponService))
	r.register(valueobject.MessageTypeWeapon, valueobject.ActionUpdateWeapon, NewWeaponHandler(r.services.WeaponService))
	r.register(valueobject.MessageTypeWeapon, valueobject.ActionDeleteWeapon, NewWeaponHandler(r.services.WeaponService))

	// User Weapon handlers
	r.register(valueobject.MessageTypeUserWeapon, valueobject.ActionGetUserWeapons, NewUserWeaponHandler(r.services.UserWeaponService))
	r.register(valueobject.MessageTypeUserWeapon, valueobject.ActionAddUserWeapon, NewUserWeaponHandler(r.services.UserWeaponService))
	r.register(valueobject.MessageTypeUserWeapon, valueobject.ActionRemoveUserWeapon, NewUserWeaponHandler(r.services.UserWeaponService))
	r.register(valueobject.MessageTypeUserWeapon, valueobject.ActionCheckUserWeapon, NewUserWeaponHandler(r.services.UserWeaponService))

	// Source Stone handlers
	r.register(valueobject.MessageTypeSourceStone, valueobject.ActionGetSourceStone, NewSourceStoneHandler(r.services.SourceStoneService))
	r.register(valueobject.MessageTypeSourceStone, valueobject.ActionGetAllSourceStones, NewSourceStoneHandler(r.services.SourceStoneService))

	// User Source Stone handlers
	r.register(valueobject.MessageTypeUserSourceStone, valueobject.ActionGetUserSourceStones, NewUserSourceStoneHandler(r.services.UserSourceStoneService))
	r.register(valueobject.MessageTypeUserSourceStone, valueobject.ActionAddUserSourceStone, NewUserSourceStoneHandler(r.services.UserSourceStoneService))
	r.register(valueobject.MessageTypeUserSourceStone, valueobject.ActionUpdateUserSourceStone, NewUserSourceStoneHandler(r.services.UserSourceStoneService))
	r.register(valueobject.MessageTypeUserSourceStone, valueobject.ActionRemoveUserSourceStone, NewUserSourceStoneHandler(r.services.UserSourceStoneService))
	r.register(valueobject.MessageTypeUserSourceStone, valueobject.ActionCheckUserSourceStone, NewUserSourceStoneHandler(r.services.UserSourceStoneService))
	r.register(valueobject.MessageTypeUserSourceStone, valueobject.ActionBatchDeleteUserSourceStone, NewUserSourceStoneHandler(r.services.UserSourceStoneService))

	// Kill Count handlers
	r.register(valueobject.MessageTypeKillCount, valueobject.ActionGetKillCount, NewKillCountHandler(r.services.KillCountService))
	r.register(valueobject.MessageTypeKillCount, valueobject.ActionGetTodayKillCount, NewKillCountHandler(r.services.KillCountService))
	r.register(valueobject.MessageTypeKillCount, valueobject.ActionUpdateKillCount, NewKillCountHandler(r.services.KillCountService))
	r.register(valueobject.MessageTypeKillCount, valueobject.ActionIncrementKillCount, NewKillCountHandler(r.services.KillCountService))
	r.register(valueobject.MessageTypeKillCount, valueobject.ActionBatchIncrementKillCount, NewKillCountHandler(r.services.KillCountService))
	r.register(valueobject.MessageTypeKillCount, valueobject.ActionGetKillRanking, NewKillCountHandler(r.services.KillCountService))
	r.register(valueobject.MessageTypeKillCount, valueobject.ActionGetUserKillRank, NewKillCountHandler(r.services.KillCountService))
	r.register(valueobject.MessageTypeKillCount, valueobject.ActionDeleteKillCount, NewKillCountHandler(r.services.KillCountService))

	// Ranking handlers
	r.register(valueobject.MessageTypeRanking, valueobject.ActionGetPlayerRanking, NewRankingHandler(r.services.RankingService))
	r.register(valueobject.MessageTypeRanking, valueobject.ActionGetUserRank, NewRankingHandler(r.services.RankingService))

	// Union handlers
	r.register(valueobject.MessageTypeUnion, valueobject.ActionGetMyUnion, NewUnionHandler(r.services.UnionService))
	r.register(valueobject.MessageTypeUnion, valueobject.ActionCreateUnion, NewUnionHandler(r.services.UnionService))
	r.register(valueobject.MessageTypeUnion, valueobject.ActionJoinUnion, NewUnionHandler(r.services.UnionService))
	r.register(valueobject.MessageTypeUnion, valueobject.ActionGetRecommendedUnions, NewUnionHandler(r.services.UnionService))
	r.register(valueobject.MessageTypeUnion, valueobject.ActionProcessUnionRequest, NewUnionHandler(r.services.UnionService))
	r.register(valueobject.MessageTypeUnion, valueobject.ActionGetUnionInfo, NewUnionHandler(r.services.UnionService))
	r.register(valueobject.MessageTypeUnion, valueobject.ActionLeaveUnion, NewUnionHandler(r.services.UnionService))
	r.register(valueobject.MessageTypeUnion, valueobject.ActionDismissUnion, NewUnionHandler(r.services.UnionService))
	r.register(valueobject.MessageTypeUnion, valueobject.ActionGetUnionRanking, NewUnionHandler(r.services.UnionService))
	r.register(valueobject.MessageTypeUnion, valueobject.ActionGetMyUnionRank, NewUnionHandler(r.services.UnionService))
	r.register(valueobject.MessageTypeUnion, valueobject.ActionGetUnionRequests, NewUnionHandler(r.services.UnionService))
	r.register(valueobject.MessageTypeUnion, valueobject.ActionInviteToUnion, NewUnionHandler(r.services.UnionService))
	r.register(valueobject.MessageTypeUnion, valueobject.ActionGetUnionInvites, NewUnionHandler(r.services.UnionService))
	r.register(valueobject.MessageTypeUnion, valueobject.ActionProcessUnionInvite, NewUnionHandler(r.services.UnionService))
}

// register registers a handler for a message type and action
func (r *messageRouter) register(msgType valueobject.MessageType, action valueobject.MessageAction, handler MessageHandler) {
	if r.handlers[msgType] == nil {
		r.handlers[msgType] = make(map[valueobject.MessageAction]MessageHandler)
	}
	r.handlers[msgType][action] = handler
	log.Printf("Registered handler for %s:%s", msgType, action)
}

// requiresAuth checks if a message type/action requires authentication
func (r *messageRouter) requiresAuth(msgType valueobject.MessageType, action valueobject.MessageAction) bool {
	// Authentication not required for these actions
	if msgType == valueobject.MessageTypeAuth && (action == valueobject.ActionLogin || action == valueobject.ActionRegister) {
		return false
	}
	if msgType == valueobject.MessageTypeHeartbeat && action == valueobject.ActionPing {
		return false
	}

	// All other actions require authentication
	return true
}