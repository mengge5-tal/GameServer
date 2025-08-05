package server

import (
	"GameServer/pkg/logger"
	"time"
)

// AuthMiddleware 认证中间件
func AuthMiddleware() Middleware {
	return MiddlewareFunc(func(c *Client, message *Message, next Handler) *Response {
		// 对于认证和心跳消息，不需要验证
		if message.Type == MessageTypeAuth || message.Type == MessageTypeHeartbeat {
			return next.Handle(c, message)
		}

		// 检查客户端是否已认证
		if !c.IsAuth {
			logger.Warn("Unauthorized access attempt", map[string]interface{}{
				"client_id":    c.ID,
				"message_type": message.Type,
				"action":       message.Action,
			})
			return NewErrorResponse(message.RequestID, CodeUnauthorized, "Authentication required")
		}

		return next.Handle(c, message)
	})
}

// LoggingMiddleware 日志中间件
func LoggingMiddleware() Middleware {
	return MiddlewareFunc(func(c *Client, message *Message, next Handler) *Response {
		start := time.Now()
		
		logger.Info("Processing message", map[string]interface{}{
			"client_id":    c.ID,
			"user_id":      c.UserID,
			"message_type": message.Type,
			"action":       message.Action,
			"request_id":   message.RequestID,
		})

		response := next.Handle(c, message)
		
		duration := time.Since(start)
		logLevel := "info"
		if !response.Success {
			logLevel = "warn"
		}

		logData := map[string]interface{}{
			"client_id":    c.ID,
			"user_id":      c.UserID,
			"message_type": message.Type,
			"action":       message.Action,
			"request_id":   message.RequestID,
			"success":      response.Success,
			"code":         response.Code,
			"duration_ms":  duration.Milliseconds(),
		}

		if logLevel == "info" {
			logger.Info("Message processed successfully", logData)
		} else {
			logger.Warn("Message processing failed", logData)
		}

		return response
	})
}

// RateLimitMiddleware 限流中间件（优化版本）
func RateLimitMiddleware() Middleware {
	return MiddlewareFunc(func(c *Client, message *Message, next Handler) *Response {
		// 使用全局限流器
		rateLimiter := GetRateLimiter()
		if rateLimiter == nil {
			// 如果限流器未初始化，跳过限流
			logger.Warn("Rate limiter not initialized, skipping rate limit check", nil)
			return next.Handle(c, message)
		}

		// 检查是否允许请求
		if !rateLimiter.IsAllowed(c.ID) {
			logger.Warn("Rate limit exceeded", map[string]interface{}{
				"client_id": c.ID,
				"user_id":   c.UserID,
				"action":    message.Action,
				"type":      message.Type,
			})
			return NewErrorResponse(message.RequestID, CodeServerError, "Rate limit exceeded")
		}

		return next.Handle(c, message)
	})
}

// ValidationMiddleware 消息验证中间件
func ValidationMiddleware() Middleware {
	return MiddlewareFunc(func(c *Client, message *Message, next Handler) *Response {
		// 验证必要字段
		if message.Type == "" {
			return NewErrorResponse(message.RequestID, CodeInvalidRequest, "Message type is required")
		}

		if message.Action == "" {
			return NewErrorResponse(message.RequestID, CodeInvalidRequest, "Action is required")
		}

		if message.RequestID == "" {
			return NewErrorResponse(message.RequestID, CodeInvalidRequest, "Request ID is required")
		}

		return next.Handle(c, message)
	})
}