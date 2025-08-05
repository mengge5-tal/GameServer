package server

import (
	"database/sql"
	"log"
)

// SetupHandlers 创建处理器适配器，将旧的处理器方法注册到路由器
func SetupHandlers(router *Router, db *sql.DB) {

	// 注册认证处理器
	router.RegisterFunc(MessageTypeAuth, ActionLogin, func(c *Client, message *Message) *Response {
		return handleLogin(c, message, db)
	})

	router.RegisterFunc(MessageTypeAuth, ActionRegister, func(c *Client, message *Message) *Response {
		return handleRegister(c, message, db)
	})

	router.RegisterFunc(MessageTypeAuth, ActionLogout, func(c *Client, message *Message) *Response {
		return handleLogout(c, message, db)
	})

	// 注册心跳处理器
	router.RegisterFunc(MessageTypeHeartbeat, ActionPing, func(c *Client, message *Message) *Response {
		return handlePing(c, message)
	})

	// 注册装备处理器
	router.RegisterFunc(MessageTypeEquip, ActionGetEquip, func(c *Client, message *Message) *Response {
		return handleGetEquip(c, message, db)
	})

	router.RegisterFunc(MessageTypeEquip, ActionSaveEquip, func(c *Client, message *Message) *Response {
		return handleSaveEquip(c, message, db)
	})

	router.RegisterFunc(MessageTypeEquip, ActionDelEquip, func(c *Client, message *Message) *Response {
		return handleDelEquip(c, message, db)
	})

	router.RegisterFunc(MessageTypeEquip, ActionBatchDelEquip, func(c *Client, message *Message) *Response {
		return handleBatchDelEquip(c, message, db)
	})

	// 注册玩家信息处理器
	router.RegisterFunc(MessageTypePlayer, ActionGetPlayerInfo, func(c *Client, message *Message) *Response {
		return handleGetPlayerInfo(c, message, db)
	})

	router.RegisterFunc(MessageTypePlayer, ActionUpdatePlayerInfo, func(c *Client, message *Message) *Response {
		return handleUpdatePlayerInfo(c, message, db)
	})

	// 注册 sourcestone 处理器
	router.RegisterFunc(MessageTypeSourcestone, "createSourcestone", func(c *Client, message *Message) *Response {
		return handleCreateSourcestone(c, message, db)
	})

	router.RegisterFunc(MessageTypeSourcestone, "getSourcestones", func(c *Client, message *Message) *Response {
		return handleGetSourcestones(c, message, db)
	})

	router.RegisterFunc(MessageTypeSourcestone, "getSourcestone", func(c *Client, message *Message) *Response {
		return handleGetSourcestone(c, message, db)
	})

	router.RegisterFunc(MessageTypeSourcestone, "updateSourcestone", func(c *Client, message *Message) *Response {
		return handleUpdateSourcestone(c, message, db)
	})

	router.RegisterFunc(MessageTypeSourcestone, "deleteSourcestone", func(c *Client, message *Message) *Response {
		return handleDeleteSourcestone(c, message, db)
	})

	router.RegisterFunc(MessageTypeSourcestone, "deleteAllSourcestones", func(c *Client, message *Message) *Response {
		return handleDeleteAllSourcestones(c, message, db)
	})

	log.Println("All handlers registered successfully")
}
