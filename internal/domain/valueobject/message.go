package valueobject

import (
	"encoding/json"
	"time"
)

// MessageType represents different types of messages
type MessageType string

const (
	MessageTypeAuth      MessageType = "auth"
	MessageTypeHeartbeat MessageType = "heartbeat"
	MessageTypeEquip     MessageType = "equip"
	MessageTypeUserEquip MessageType = "userequip"
	MessageTypePlayer    MessageType = "player"
	MessageTypeFriend    MessageType = "friend"
	MessageTypeRank      MessageType = "rank"
	MessageTypeOnline    MessageType = "online"
)

// MessageAction represents different actions within message types
type MessageAction string

const (
	// Auth actions
	ActionLogin    MessageAction = "login"
	ActionRegister MessageAction = "register"
	ActionLogout   MessageAction = "logout"

	// Heartbeat actions
	ActionPing MessageAction = "ping"

	// Equipment actions
	ActionGetEquip    MessageAction = "getEquip"
	ActionSaveEquip   MessageAction = "saveEquip"
	ActionDeleteEquip MessageAction = "deleteEquip"
	ActionDelEquip    MessageAction = "delEquip"

	// User Equipment actions
	ActionGetEquippedItems  MessageAction = "getEquippedItems"
	ActionEquipItem         MessageAction = "equipItem"
	ActionUnequipItem       MessageAction = "unequipItem"
	ActionGetEquipmentStats MessageAction = "getEquipmentStats"
	ActionGetEquippedBySlot MessageAction = "getEquippedBySlot"

	// Player actions
	ActionGetPlayerInfo MessageAction = "getPlayerInfo"
	ActionUpdatePlayer  MessageAction = "updatePlayer"

	// Friend actions
	ActionGetFriends       MessageAction = "getFriends"
	ActionAddFriend        MessageAction = "addFriend"
	ActionRemoveFriend     MessageAction = "removeFriend"
	ActionAcceptFriend     MessageAction = "acceptFriend"
	ActionRejectFriend     MessageAction = "rejectFriend"
	ActionGetFriendRank    MessageAction = "getFriendRank"

	// Rank actions
	ActionGetAllRank MessageAction = "getAllRank"
	ActionGetRank    MessageAction = "getRank"

	// Online actions
	ActionGetOnlineUsers MessageAction = "getOnlineUsers"
)

// Message represents a WebSocket message
type Message struct {
	Type      MessageType     `json:"type"`
	Action    MessageAction   `json:"action"`
	Data      json.RawMessage `json:"data"`
	RequestID string          `json:"requestId"`
	Timestamp int64           `json:"timestamp"`
}

// Response represents a WebSocket response
type Response struct {
	Success   bool        `json:"success"`
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data"`
	RequestID string      `json:"requestId"`
	Timestamp int64       `json:"timestamp"`
}

// ResponseCode defines response codes
type ResponseCode int

const (
	CodeSuccess        ResponseCode = 0
	CodeInvalidRequest ResponseCode = 1001
	CodeUnauthorized   ResponseCode = 1002
	CodeForbidden      ResponseCode = 1003
	CodeNotFound       ResponseCode = 1004
	CodeConflict       ResponseCode = 1005
	CodeValidationError ResponseCode = 1006
	CodeInternalError  ResponseCode = 5000
)

// NewSuccessResponse creates a success response
func NewSuccessResponse(requestID string, data interface{}) *Response {
	return &Response{
		Success:   true,
		Code:      int(CodeSuccess),
		Message:   "Success",
		Data:      data,
		RequestID: requestID,
		Timestamp: time.Now().Unix(),
	}
}

// NewErrorResponse creates an error response
func NewErrorResponse(requestID string, code ResponseCode, message string) *Response {
	return &Response{
		Success:   false,
		Code:      int(code),
		Message:   message,
		Data:      nil,
		RequestID: requestID,
		Timestamp: time.Now().Unix(),
	}
}

// ToJSON converts response to JSON bytes
func (r *Response) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

// ParseMessage parses JSON bytes to Message
func ParseMessage(data []byte) (*Message, error) {
	var msg Message
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}