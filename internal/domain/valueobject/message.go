package valueobject

import (
	"encoding/json"
	"time"
	"GameServer/pkg/utils"
)

// MessageType represents different types of messages
type MessageType string

const (
	MessageTypeAuth       MessageType = "auth"
	MessageTypeHeartbeat  MessageType = "heartbeat"
	MessageTypeEquip      MessageType = "equip"
	MessageTypeUserEquip  MessageType = "userequip"
	MessageTypePlayer     MessageType = "player"
	MessageTypeFriend     MessageType = "friend"
	MessageTypeOnline     MessageType = "online"
	MessageTypeExperience MessageType = "experience"
	MessageTypeWeapon          MessageType = "weapon"
	MessageTypeUserWeapon      MessageType = "userweapon"
	MessageTypeSourceStone     MessageType = "sourcestone"
	MessageTypeUserSourceStone MessageType = "usersourcestone"
	MessageTypeKillCount       MessageType = "killcount"
	MessageTypeRanking         MessageType = "ranking"
	MessageTypeUnion           MessageType = "union"
	MessageTypeChat            MessageType = "chat"
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
	ActionBatchDeleteEquip MessageAction = "batchDeleteEquip"

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
	ActionFriendRequest    MessageAction = "friendRequest"
	ActionFriendResponse   MessageAction = "friendResponse"
	ActionGetRecommendedFriends MessageAction = "getRecommendedFriends"
	
	// Friend notification (server-sent)
	ActionFriendRequestNotification MessageAction = "friendRequestNotification"


	// Online actions
	ActionGetOnlineUsers MessageAction = "getOnlineUsers"

	// Experience actions
	ActionGetByLevel   MessageAction = "getByLevel"
	ActionGetAllLevels MessageAction = "getAllLevels"

	// Weapon actions
	ActionGetWeapon     MessageAction = "getWeapon"
	ActionGetAllWeapons MessageAction = "getAllWeapons"
	ActionCreateWeapon  MessageAction = "createWeapon"
	ActionUpdateWeapon  MessageAction = "updateWeapon"
	ActionDeleteWeapon  MessageAction = "deleteWeapon"

	// User Weapon actions
	ActionGetUserWeapons    MessageAction = "getUserWeapons"
	ActionAddUserWeapon     MessageAction = "addUserWeapon"
	ActionRemoveUserWeapon  MessageAction = "removeUserWeapon"
	ActionCheckUserWeapon   MessageAction = "checkUserWeapon"

	// Source Stone actions
	ActionGetSourceStone     MessageAction = "getSourceStone"
	ActionGetAllSourceStones MessageAction = "getAllSourceStones"

	// User Source Stone actions
	ActionGetUserSourceStones    MessageAction = "getUserSourceStones"
	ActionAddUserSourceStone     MessageAction = "addUserSourceStone"
	ActionUpdateUserSourceStone  MessageAction = "updateUserSourceStone"
	ActionRemoveUserSourceStone  MessageAction = "removeUserSourceStone"
	ActionCheckUserSourceStone   MessageAction = "checkUserSourceStone"
	ActionBatchDeleteUserSourceStone MessageAction = "batchDeleteUserSourceStone"

	// Kill Count actions
	ActionGetKillCount            MessageAction = "getKillCount"
	ActionGetTodayKillCount       MessageAction = "getTodayKillCount"
	ActionUpdateKillCount         MessageAction = "updateKillCount"
	ActionIncrementKillCount      MessageAction = "incrementKillCount"
	ActionBatchIncrementKillCount MessageAction = "batchIncrementKillCount"
	ActionDeleteKillCount         MessageAction = "deleteKillCount"
	ActionGetKillRanking          MessageAction = "getKillRanking"
	ActionGetUserKillRank         MessageAction = "getUserKillRank"

	// Ranking actions
	ActionGetPlayerRanking MessageAction = "getPlayerRanking"
	ActionGetUserRank      MessageAction = "getUserRank"

	// Union actions
	ActionCreateUnion          MessageAction = "createUnion"
	ActionJoinUnion            MessageAction = "joinUnion"
	ActionLeaveUnion           MessageAction = "leaveUnion"
	ActionGetMyUnion           MessageAction = "getMyUnion"
	ActionGetUnionInfo         MessageAction = "getUnionInfo"
	ActionGetRecommendedUnions MessageAction = "getRecommendedUnions"
	ActionProcessUnionRequest  MessageAction = "processUnionRequest"
	ActionGetUnionRanking      MessageAction = "getUnionRanking"
	ActionGetMyUnionRank       MessageAction = "getMyUnionRank"
	ActionDismissUnion         MessageAction = "dismissUnion"
	ActionGetUnionRequests     MessageAction = "getUnionRequests"
	ActionInviteToUnion        MessageAction = "inviteToUnion"
	ActionGetUnionInvites      MessageAction = "getUnionInvites"
	ActionProcessUnionInvite   MessageAction = "processUnionInvite"
	ActionPromoteMember        MessageAction = "promoteMember"
	ActionDemoteMember         MessageAction = "demoteMember"
	ActionKickMember           MessageAction = "kickMember"
	ActionTransferLeadership   MessageAction = "transferLeadership"
	ActionGetUnionMembers      MessageAction = "getUnionMembers"
	ActionSearchUnionMembers   MessageAction = "searchUnionMembers"
	ActionUpdateUnionInfo      MessageAction = "updateUnionInfo"

	// Chat actions
	ActionSendPrivateMessage      MessageAction = "sendPrivateMessage"
	ActionGetPrivateMessages      MessageAction = "getPrivateMessages"
	ActionSendWorldMessage        MessageAction = "sendWorldMessage"
	ActionJoinWorldChannel        MessageAction = "joinWorldChannel"
	ActionLeaveWorldChannel       MessageAction = "leaveWorldChannel"
	ActionGetWorldChannels        MessageAction = "getWorldChannels"
	ActionSendUnionMessage        MessageAction = "sendUnionMessage"
	ActionGetUnionMessages        MessageAction = "getUnionMessages"
	ActionGetRecentUnionMessages  MessageAction = "getRecentUnionMessages"
	ActionGetChatStats            MessageAction = "getChatStats"
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

var requestIDGenerator = utils.NewRequestIDGenerator()

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

// NewSuccessResponseWithUniqueID creates a success response with unique request ID
func NewSuccessResponseWithUniqueID(msgType MessageType, action MessageAction, data interface{}) *Response {
	uniqueID := requestIDGenerator.GenerateSimple(string(msgType), string(action))
	return &Response{
		Success:   true,
		Code:      int(CodeSuccess),
		Message:   "Success",
		Data:      data,
		RequestID: uniqueID,
		Timestamp: time.Now().Unix(),
	}
}

// NewErrorResponseWithUniqueID creates an error response with unique request ID
func NewErrorResponseWithUniqueID(msgType MessageType, action MessageAction, code ResponseCode, message string) *Response {
	uniqueID := requestIDGenerator.GenerateSimple(string(msgType), string(action))
	return &Response{
		Success:   false,
		Code:      int(code),
		Message:   message,
		Data:      nil,
		RequestID: uniqueID,
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