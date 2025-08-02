package server

import (
	"encoding/json"
	"time"
)

// 统一的消息格式
type Message struct {
	Type      string      `json:"type"`      // 消息类型
	Action    string      `json:"action"`    // 具体操作
	Data      interface{} `json:"data"`      // 数据内容
	RequestID string      `json:"requestId"` // 请求ID，用于匹配请求和响应
	Timestamp int64       `json:"timestamp"` // 时间戳
}

// 统一的响应格式
type Response struct {
	Success   bool        `json:"success"`   // 是否成功
	Code      int         `json:"code"`      // 错误码
	Message   string      `json:"message"`   // 错误信息
	Data      interface{} `json:"data"`      // 返回数据
	RequestID string      `json:"requestId"` // 对应的请求ID
	Timestamp int64       `json:"timestamp"` // 时间戳
}

// 错误码定义
const (
	CodeSuccess          = 0    // 成功
	CodeInvalidRequest   = 1001 // 无效请求
	CodeUserNotFound     = 1002 // 用户不存在
	CodeWrongPassword    = 1003 // 密码错误
	CodeUserExists       = 1004 // 用户已存在
	CodeDatabaseError    = 1005 // 数据库错误
	CodeUnauthorized     = 1006 // 未授权
	CodeInvalidParams    = 1007 // 参数错误
	CodeServerError      = 1008 // 服务器内部错误
)

// 消息类型定义
const (
	MessageTypeAuth     = "auth"      // 认证相关
	MessageTypeUser     = "user"      // 用户相关
	MessageTypeEquip    = "equip"     // 装备相关
	MessageTypePlayer   = "player"    // 玩家信息相关
	MessageTypeFriend   = "friend"    // 好友相关
	MessageTypeRank     = "rank"      // 排行榜相关
	MessageTypeHeartbeat = "heartbeat" // 心跳
)

// 操作类型定义
const (
	ActionLogin    = "login"
	ActionRegister = "register"
	ActionLogout   = "logout"
	
	ActionGetEquip     = "getEquip"
	ActionSaveEquip    = "saveEquip"
	ActionDelEquip     = "delEquip"
	ActionBatchDelEquip = "batchDelEquip"
	
	ActionGetPlayerInfo    = "getPlayerInfo"
	ActionUpdatePlayerInfo = "updatePlayerInfo"
	
	ActionGetFriends      = "getFriends"
	ActionUpdateFriends   = "updateFriends"
	ActionAddFriend       = "addFriend"
	ActionDelFriend       = "delFriend"
	ActionFriendRequest   = "friendRequest"
	ActionFriendResponse  = "friendResponse"
	
	ActionGetAllRank = "getAllRank"
	ActionGetSelfRank = "getSelfRank"
	
	ActionPing = "ping"
	ActionPong = "pong"
)

// 创建成功响应
func NewSuccessResponse(requestID string, data interface{}) *Response {
	return &Response{
		Success:   true,
		Code:      CodeSuccess,
		Message:   "Success",
		Data:      data,
		RequestID: requestID,
		Timestamp: time.Now().Unix(),
	}
}

// 创建错误响应
func NewErrorResponse(requestID string, code int, message string) *Response {
	return &Response{
		Success:   false,
		Code:      code,
		Message:   message,
		Data:      nil,
		RequestID: requestID,
		Timestamp: time.Now().Unix(),
	}
}

// 将消息转换为JSON字节数组
func (m *Message) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}

// 将响应转换为JSON字节数组
func (r *Response) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

// 从JSON字节数组解析消息
func ParseMessage(data []byte) (*Message, error) {
	var msg Message
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}