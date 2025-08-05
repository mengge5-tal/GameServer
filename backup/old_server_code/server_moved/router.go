package server

import (
	"fmt"
	"log"
)

// Handler 处理器接口
type Handler interface {
	Handle(c *Client, message *Message) *Response
}

// HandlerFunc 处理器函数类型
type HandlerFunc func(c *Client, message *Message) *Response

// Handle 实现Handler接口
func (f HandlerFunc) Handle(c *Client, message *Message) *Response {
	return f(c, message)
}

// Middleware 中间件接口
type Middleware interface {
	Process(c *Client, message *Message, next Handler) *Response
}

// MiddlewareFunc 中间件函数类型
type MiddlewareFunc func(c *Client, message *Message, next Handler) *Response

// Process 实现Middleware接口
func (f MiddlewareFunc) Process(c *Client, message *Message, next Handler) *Response {
	return f(c, message, next)
}

// Router 路由器
type Router struct {
	handlers    map[string]map[string]Handler
	middlewares []Middleware
}

// NewRouter 创建新的路由器
func NewRouter() *Router {
	return &Router{
		handlers:    make(map[string]map[string]Handler),
		middlewares: make([]Middleware, 0),
	}
}

// Use 添加中间件
func (r *Router) Use(middleware Middleware) {
	r.middlewares = append(r.middlewares, middleware)
}

// Register 注册处理器
func (r *Router) Register(messageType, action string, handler Handler) {
	if r.handlers[messageType] == nil {
		r.handlers[messageType] = make(map[string]Handler)
	}
	r.handlers[messageType][action] = handler
	log.Printf("Registered handler for %s:%s", messageType, action)
}

// RegisterFunc 注册处理器函数
func (r *Router) RegisterFunc(messageType, action string, handlerFunc HandlerFunc) {
	r.Register(messageType, action, handlerFunc)
}

// Handle 处理消息
func (r *Router) Handle(c *Client, message *Message) *Response {
	// 查找处理器
	typeHandlers, exists := r.handlers[message.Type]
	if !exists {
		log.Printf("Unknown message type: %s", message.Type)
		return NewErrorResponse(message.RequestID, CodeInvalidRequest, "Unknown message type")
	}

	handler, exists := typeHandlers[message.Action]
	if !exists {
		log.Printf("Unknown action %s for type %s", message.Action, message.Type)
		return NewErrorResponse(message.RequestID, CodeInvalidRequest, 
			fmt.Sprintf("Unknown action %s for type %s", message.Action, message.Type))
	}

	// 构建中间件链
	finalHandler := handler
	for i := len(r.middlewares) - 1; i >= 0; i-- {
		middleware := r.middlewares[i]
		nextHandler := finalHandler
		finalHandler = HandlerFunc(func(c *Client, message *Message) *Response {
			return middleware.Process(c, message, nextHandler)
		})
	}

	// 执行处理器
	return finalHandler.Handle(c, message)
}

// GetRoutes 获取所有已注册的路由信息
func (r *Router) GetRoutes() map[string][]string {
	routes := make(map[string][]string)
	for messageType, actions := range r.handlers {
		for action := range actions {
			routes[messageType] = append(routes[messageType], action)
		}
	}
	return routes
}