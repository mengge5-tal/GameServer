package types

import "database/sql"

// Client 客户端接口定义
type Client interface {
	GetID() string
	GetUserID() int
	SetAuth(bool)
	SetUserID(int)
	GetHub() Hub
	SendResponse(response Response)
}

// Hub 连接管理器接口定义
type Hub interface {
	GetDB() *sql.DB
	SetUserClient(userID int, client Client)
	GetClientByUserID(userID int) Client
	RemoveUserClient(userID int)
}

// Message 消息接口定义
type Message interface {
	GetType() string
	GetAction() string
	GetData() interface{}
	GetRequestID() string
	GetTimestamp() int64
}

// Response 响应接口定义
type Response interface {
	GetSuccess() bool
	GetCode() int
	GetMessage() string
	GetData() interface{}
	GetRequestID() string
	GetTimestamp() int64
	ToJSON() ([]byte, error)
}