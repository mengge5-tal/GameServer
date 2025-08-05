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
		return valueobject.NewErrorResponse(message.RequestID, valueobject.CodeInvalidRequest, "Unknown message type")
	}

	// Find handler for action
	handler, exists := typeHandlers[message.Action]
	if !exists {
		log.Printf("Unknown action %s for type %s", message.Action, message.Type)
		return valueobject.NewErrorResponse(message.RequestID, valueobject.CodeInvalidRequest, "Unknown action")
	}

	// Check authentication for protected actions
	if r.requiresAuth(message.Type, message.Action) && !client.IsAuthenticated() {
		return valueobject.NewErrorResponse(message.RequestID, valueobject.CodeUnauthorized, "Authentication required")
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

	// Friend handlers
	r.register(valueobject.MessageTypeFriend, valueobject.ActionGetFriends, NewFriendHandler(r.services.FriendService))
	r.register(valueobject.MessageTypeFriend, valueobject.ActionAddFriend, NewFriendHandler(r.services.FriendService))
	r.register(valueobject.MessageTypeFriend, valueobject.ActionRemoveFriend, NewFriendHandler(r.services.FriendService))
	r.register(valueobject.MessageTypeFriend, valueobject.ActionAcceptFriend, NewFriendHandler(r.services.FriendService))
	r.register(valueobject.MessageTypeFriend, valueobject.ActionRejectFriend, NewFriendHandler(r.services.FriendService))
	r.register(valueobject.MessageTypeFriend, valueobject.ActionGetFriendRank, NewFriendHandler(r.services.FriendService))

	// Ranking handlers
	r.register(valueobject.MessageTypeRank, valueobject.ActionGetAllRank, NewRankingHandler(r.services.RankingService))
	r.register(valueobject.MessageTypeRank, valueobject.ActionGetRank, NewRankingHandler(r.services.RankingService))
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