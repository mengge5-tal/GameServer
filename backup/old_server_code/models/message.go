package models

import (
	"encoding/json"
	"fmt"
)

// Message types
const (
	MessageTypeAuth      = "auth"
	MessageTypeHeartbeat = "heartbeat"
	MessageTypeUser      = "user"
	MessageTypeEquip     = "equip"
	MessageTypePlayer    = "player"
	MessageTypeFriend    = "friend"
	MessageTypeRank      = "rank"
)

// Actions
const (
	// Auth actions
	ActionLogin    = "login"
	ActionRegister = "register"
	ActionLogout   = "logout"

	// Heartbeat actions
	ActionPing = "ping"
	ActionPong = "pong"

	// Equipment actions
	ActionGetEquip      = "getEquip"
	ActionSaveEquip     = "saveEquip"
	ActionDelEquip      = "delEquip"
	ActionBatchDelEquip = "batchDelEquip"

	// Player actions
	ActionGetPlayerInfo    = "getPlayerInfo"
	ActionUpdatePlayerInfo = "updatePlayerInfo"

	// Friend actions
	ActionSendFriendRequest    = "sendFriendRequest"
	ActionAcceptFriendRequest  = "acceptFriendRequest"
	ActionRejectFriendRequest  = "rejectFriendRequest"
	ActionGetFriendRequests    = "getFriendRequests"
	ActionGetFriendsList       = "getFriendsList"
	ActionRemoveFriend         = "removeFriend"

	// Rank actions
	ActionGetRanking       = "getRanking"
	ActionUpdateRanking    = "updateRanking"
)

// Response codes
const (
	CodeSuccess         = 200
	CodeInvalidRequest  = 400
	CodeUnauthorized    = 401
	CodeUserNotFound    = 404
	CodeUserExists      = 409
	CodeInvalidParams   = 422
	CodeWrongPassword   = 423
	CodeServerError     = 500
	CodeDatabaseError   = 501
)

// Message represents a WebSocket message
type Message struct {
	Type      string      `json:"type"`
	Action    string      `json:"action"`
	Data      interface{} `json:"data"`
	RequestID string      `json:"requestId"`
	Timestamp int64       `json:"timestamp"`
}

// Response represents a WebSocket response
type Response struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data"`
	RequestID string      `json:"requestId"`
	Timestamp int64       `json:"timestamp"`
}

// ParseMessage parses JSON data into a Message struct
func ParseMessage(data []byte) (*Message, error) {
	var message Message
	if err := json.Unmarshal(data, &message); err != nil {
		return nil, fmt.Errorf("failed to parse message: %w", err)
	}
	return &message, nil
}

// ToJSON converts response to JSON bytes
func (r *Response) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

// NewSuccessResponse creates a successful response
func NewSuccessResponse(requestID string, data interface{}) *Response {
	return &Response{
		Code:      CodeSuccess,
		Message:   "Success",
		Data:      data,
		RequestID: requestID,
		Timestamp: GetCurrentTimestamp(),
	}
}

// NewErrorResponse creates an error response
func NewErrorResponse(requestID string, code int, message string) *Response {
	return &Response{
		Code:      code,
		Message:   message,
		Data:      nil,
		RequestID: requestID,
		Timestamp: GetCurrentTimestamp(),
	}
}

// GetCurrentTimestamp returns current timestamp
func GetCurrentTimestamp() int64 {
	return 1640995200 // You can implement proper timestamp here
}